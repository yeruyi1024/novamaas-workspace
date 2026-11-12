"""Publication boundaries and current-HEAD validation for upstream sync CI."""

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import unittest
from unittest import mock

import test_upstream_sync as fixtures
import test_upstream_sync_handoff as handoff_fixtures
import upstream_sync_ci as ci


class UpstreamSyncPublicationTests(unittest.TestCase):
    def setUp(self):
        self.fixture = fixtures.UpstreamSyncIntegrationTests()
        self.fixture.setUp()
        self.addCleanup(self.fixture.doCleanups)
        f = self.fixture
        self.env = mock.patch.dict(os.environ, dict(f.env, GITHUB_RUN_ID="12345", GITHUB_RUN_ATTEMPT="1",
            DEFAULT_BRANCH="main", GIT_AUTHOR_NAME="Fixture Bot", GIT_AUTHOR_EMAIL="bot@example.invalid",
            GIT_COMMITTER_NAME="Fixture Bot", GIT_COMMITTER_EMAIL="bot@example.invalid"))
        self.env.start()
        self.addCleanup(self.env.stop)
        for name in ("upstream_sync.py", "upstream_sync_handoff.py", "upstream_sync_handoff_context.py"):
            f.write(".github/scripts/" + name, Path(__file__).with_name(name).read_text())
        f.write(".gitignore", Path(__file__).resolve().parents[2].joinpath(".gitignore").read_text())
        f.commit("Add trusted sync executor")
        f.base = f.git("rev-parse", "HEAD").strip()
        f.git("remote", "add", "origin", str(f.repo))
        self.output = f.root / "publication"
        self.args = argparse.Namespace(source=f.repo, output=self.output, dry_run=True)

    def make_candidate(self):
        f = self.fixture
        f.reviewed_plan()
        f.plan["items"][0]["decision"] = "merge"
        f.plan["items"][0]["reason"] = "Selected provider regression fix."
        f.save_plan()
        plan_checkout = f.root / "planning"
        f.git("worktree", "add", "-b", "codex/upstream-plan-fixture", str(plan_checkout), f.base)
        relative = "docs/upstream-sync/plans/fixture-20261016/plan.json"
        f.write(relative, f.plan_path.read_text(), root=plan_checkout)
        f.commit("Review upstream plan", repo=plan_checkout)
        os.environ.update(PLAN_REF="codex/upstream-plan-fixture", PLAN_PATH=relative)
        f.apply()
        verification = f.verification()
        f.finalize(verification)
        ci.bundle(f.candidate, self.output, "apply", f.plan["batch_id"], f.base, f.plan["source"]["tag"])
        shutil.copyfile(f.result_path(), self.output / "result.json")
        shutil.copyfile(verification, self.output / "verification.json")

    def test_candidate_bundle_preserves_verified_commit_chain_without_push(self):
        self.make_candidate()
        f = self.fixture
        before = f.git("rev-parse", "HEAD").strip()
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "api") as api:
            ci.publish(self.args)
            api.assert_not_called()
        manifest = json.loads((self.output / "publication.json").read_text())
        self.assertEqual(f.git("rev-parse", "refs/upstream-sync/candidate").strip(), manifest["candidate_head"])
        self.assertEqual(f.git("rev-parse", "HEAD").strip(), before)
        self.assertEqual(f.git("status", "--porcelain").strip(), "")
        self.assertEqual((f.repo / "provider.txt").read_text(), "existing provider behavior\n")
        history = json.loads((f.candidate / "docs/upstream-sync/history/fixture-20261016.json").read_text())
        mapped = history["mappings"][0]["candidate_sha"]
        f.git("merge-base", "--is-ancestor", mapped, manifest["candidate_head"])

    def test_publisher_rejects_changed_reviewed_plan(self):
        self.make_candidate()
        f = self.fixture
        planning = f.root / "planning"
        path = planning / os.environ["PLAN_PATH"]
        plan = json.loads(path.read_text())
        plan["items"][0]["reason"] = "Changed after execution."
        path.write_text(json.dumps(plan))
        f.commit("Amend reviewed decision", repo=planning)
        with mock.patch.object(ci, "fetch_upstream"), self.assertRaisesRegex(ValueError, "reviewed planning branch"):
            ci.publish(self.args)

    def test_publisher_rejects_another_run_artifact(self):
        self.make_candidate()
        path = self.output / "publication.json"
        manifest = json.loads(path.read_text())
        manifest["run_id"] = "67890"
        path.write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, "current workflow run"):
            ci.publish(self.args)

    def test_publisher_rejects_candidate_bundle_with_different_tree(self):
        self.make_candidate()
        f = self.fixture
        f.write("provider.txt", "unreviewed mutation\n", root=f.candidate)
        f.commit("Unreviewed change after validation", repo=f.candidate)
        manifest_path = self.output / "publication.json"
        manifest = json.loads(manifest_path.read_text())
        manifest["candidate_head"] = f.git("rev-parse", "HEAD", repo=f.candidate).strip()
        manifest_path.write_text(json.dumps(manifest))
        ci.git(f.candidate, "bundle", "create", str(self.output / "candidate.bundle"), "HEAD", "^" + f.base)
        with self.assertRaisesRegex(ValueError, "different trees"):
            ci.publish(self.args)

    def test_assessment_patch_opens_only_plan_files_from_fresh_base(self):
        f = self.fixture
        self.assertEqual(f.git("check-ignore", "plans/local-only.json", check=False).returncode, 0)
        os.environ.update(TARGET="v1.1.0", RANGE_BASE=f.baseline, BATCH="fresh-assessment",
                          NEXT_VERSION="v1.0.0-rc.26-novamaas.3")
        with mock.patch.object(ci, "fetch_upstream"):
            ci.analyze(argparse.Namespace(source=f.repo, trusted=f.repo, output=self.output))
        publisher = f.root / "publisher"
        f.git("clone", str(f.repo), str(publisher))
        ci.publish(argparse.Namespace(source=publisher, output=self.output, dry_run=True))
        names = f.git("diff", "--cached", "--name-only", repo=publisher).splitlines()
        self.assertEqual(set(names), {"docs/upstream-sync/plans/fresh-assessment/plan.json",
                                     "docs/upstream-sync/plans/fresh-assessment/report.md",
                                     "docs/upstream-sync/plans/fresh-assessment/AI_HANDOFF.md"})
        self.assertEqual(f.git("rev-parse", "HEAD", repo=publisher).strip(), f.base)
        self.assertFalse((publisher / "docs/upstream-sync/plans/fresh-assessment").exists())

    def test_current_candidate_head_is_revalidated_using_trusted_records(self):
        self.make_candidate()
        f = self.fixture
        head = f.git("rev-parse", "HEAD", repo=f.candidate).strip()
        os.environ.update(PR_BASE_SHA=f.base, PR_HEAD_SHA=head, PR_HEAD_REF="codex/upstream-apply-fixture")
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "verify_candidate") as verify:
            ci.revalidate(argparse.Namespace(candidate=f.candidate, trusted=f.repo, output=self.output))
            verify.assert_called_once_with(f.candidate.resolve(), self.output.resolve() / "verification.json")

    def test_generated_branch_cannot_remove_history_to_skip_validation(self):
        f = self.fixture
        os.environ.update(PR_BASE_SHA=f.base, PR_HEAD_SHA=f.base, PR_HEAD_REF="codex/upstream-apply-fixture")
        with self.assertRaisesRegex(ValueError, "retain its synchronization history"):
            ci.revalidate(argparse.Namespace(candidate=f.repo, trusted=f.repo, output=self.output))

    def manual_pr(self):
        handoff = handoff_fixtures.UpstreamSyncHandoffTests()
        handoff.setUp()
        self.addCleanup(handoff.doCleanups)
        os.environ.update(PR_BASE_SHA=handoff.fixture.base, PR_HEAD_SHA=handoff.head(),
                          PR_HEAD_REF="codex/upstream-plan-fixture", PR_DRAFT="true")
        return handoff

    def test_planning_draft_retains_pending_state_and_skips_candidate_builds(self):
        handoff = self.manual_pr()
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "verify_candidate") as verify:
            ci.revalidate(argparse.Namespace(candidate=handoff.fixture.repo, trusted=self.fixture.repo,
                                             output=self.output))
            verify.assert_not_called()
        evidence = json.loads((self.output / "verification.json").read_text())
        self.assertEqual(evidence["status"], "planning")
        self.assertEqual(evidence["validated_head"], handoff.head())
        self.assertEqual(evidence["checks"], [])
        self.assertFalse(handoff.history_path.exists())

    def test_planning_only_pr_cannot_become_ready_without_implementation(self):
        handoff = self.manual_pr()
        os.environ["PR_DRAFT"] = "false"
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "verify_candidate") as verify:
            with self.assertRaisesRegex(ValueError, "must remain Draft"):
                ci.revalidate(argparse.Namespace(candidate=handoff.fixture.repo, trusted=self.fixture.repo,
                                                 output=self.output))
            verify.assert_not_called()
        self.assertFalse((self.output / "verification.json").exists())

    def test_manual_handoff_pr_revalidates_actual_current_head_after_source_proof(self):
        handoff = self.manual_pr()
        handoff.approve_plan()
        handoff.write_runtime()
        handoff.finalize()
        os.environ.update(PR_HEAD_SHA=handoff.head(), PR_DRAFT="false")
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "verify_candidate") as verify:
            ci.revalidate(argparse.Namespace(candidate=handoff.fixture.repo, trusted=self.fixture.repo,
                                             output=self.output))
            verify.assert_called_once_with(handoff.fixture.repo.resolve(), self.output.resolve() / "verification.json")
        self.assertEqual(json.loads(handoff.history_path.read_text())["state"], "ready_for_ci")

    def test_stale_manual_handoff_record_cannot_reach_builds_or_claim_passing_checks(self):
        handoff = self.manual_pr()
        handoff.approve_plan()
        handoff.write_runtime()
        handoff.finalize()
        handoff.write_runtime("Appended regression fix has not been recorded for CI yet\n")
        os.environ["PR_HEAD_SHA"] = handoff.head()
        with mock.patch.object(ci, "fetch_upstream"), mock.patch.object(ci, "verify_candidate") as verify:
            with self.assertRaises(subprocess.CalledProcessError):
                ci.revalidate(argparse.Namespace(candidate=handoff.fixture.repo, trusted=self.fixture.repo,
                                                 output=self.output))
            verify.assert_not_called()
        self.assertFalse((self.output / "verification.json").exists())

    def test_private_origin_auth_is_temporary_and_excluded_from_upstream_fetch(self):
        f = self.fixture
        os.environ["GH_TOKEN"] = "fixture-write-token"
        with mock.patch.object(ci, "git") as call:
            ci.origin(f.repo, "ls-remote", "--heads", "origin", "refs/heads/main")
            environment = call.call_args.kwargs["env"]
            self.assertEqual(environment["GIT_CONFIG_KEY_0"], "http.extraheader")
            self.assertTrue(environment["GIT_CONFIG_VALUE_0"].startswith("AUTHORIZATION: basic "))
            self.assertNotIn("GIT_CONFIG_VALUE_0", os.environ)
            call.reset_mock()
            ci.fetch_upstream(f.repo, "v1.1.0", f.baseline)
            self.assertEqual(len(call.call_args_list), 2)
            self.assertTrue(all("env" not in entry.kwargs for entry in call.call_args_list))

    def test_validation_child_has_no_github_credentials_or_runner_command_files(self):
        f = self.fixture
        os.environ.update(GH_TOKEN="fixture-secret", GITHUB_TOKEN="fixture-secret", GITHUB_ENV="/tmp/runner-env",
                          GITHUB_OUTPUT="/tmp/runner-output", ACTIONS_RUNTIME_TOKEN="fixture-runtime")
        calls = []
        def capture(command, cwd=None, capture=False, env=None):
            calls.append((command, env))
        with mock.patch.object(ci, "run", side_effect=capture), mock.patch.object(ci, "git", side_effect=[f.base, f.base, ""]):
            ci.verify_candidate(f.repo, self.output / "verification.json")
        self.assertEqual(len(calls), 8)
        for _, env in calls:
            self.assertTrue(all(not key.startswith(("GITHUB_", "ACTIONS_", "GH_")) for key in env))
        checks = json.loads((self.output / "verification.json").read_text())["checks"]
        self.assertEqual({entry["name"] for entry in checks}, set(ci.CHECKS))


if __name__ == "__main__":
    unittest.main()
