# Upstream synchronization assessment: rc44-assessment-20261016

This report contains mechanical evidence, any recorded AI advice and review prompts. Advice does not establish semantic compatibility or authorize implementation.

- Source: https://github.com/QuantumNous/new-api
- Target release: `v1.0.0-rc.44` (`633b4c5b190dc1af99a2814391354d86948e7943`)
- Evaluated range: `8f6961c675932f406260ff0c218bc2aa0603e9b2..633b4c5b190dc1af99a2814391354d86948e7943`
- Preserved initial fork baseline: `8f6961c675932f406260ff0c218bc2aa0603e9b2`
- Downstream base: `b2cb1c2473f1db803ff5bdaafc67699288bde713` (`v1.0.0-rc.26-novamaas.2`)
- Proposed downstream version: `v1.0.0-rc.26-novamaas.3`

Each source commit is a review item, not a proven feature boundary. Use feature/group names and dependencies to describe cross-commit work.
All decisions need a reason. `covered` also needs evidence; patch equivalence and old source references are evidence only.
`merge` and `upgrade` use source commits; approved AI scopes permit conflict adaptation. `migrate` uses a reviewed patch or an explicitly approved AI scope; no commands run from the plan.
Sensitive CI/policy/license/provenance/version paths, symlinks, submodules and changes to existing attribution lines cannot be applied automatically.

## Historical follow-up outside this evaluated range

Previous skipped/deferred sources and partial migrations stay visible here. These are review evidence, not selected work; a later full adoption or documented coverage supersedes them.

No historical follow-up outside this range is recorded.

## Current evaluated range

| Item | Source | Decision | Files | Downstream overlap | Review signals |
| --- | --- | --- | ---: | ---: | --- |
| commit-cae3676ec6f4 | [cae3676ec6f4](https://github.com/QuantumNous/new-api/commit/cae3676ec6f46ee5ef596443256f78c4e9b34ceb) | pending | 1 | 0 | review required |
| commit-ba2e9287bb7a | [ba2e9287bb7a](https://github.com/QuantumNous/new-api/commit/ba2e9287bb7a8002116c03daa4c457a330054871) | pending | 1 | 0 | review required |
| commit-e468b73915e5 | [e468b73915e5](https://github.com/QuantumNous/new-api/commit/e468b73915e5028e9849de62c5018a0faa203012) | pending | 10 | 3 | protected files |
| commit-692e8d6ee6a9 | [692e8d6ee6a9](https://github.com/QuantumNous/new-api/commit/692e8d6ee6a9a1620c2d731cb51a1e3154a7042b) | pending | 2 | 1 | review required |
| commit-ac381acf4bf4 | [ac381acf4bf4](https://github.com/QuantumNous/new-api/commit/ac381acf4bf41204b97bb26b4c58c83275877a2e) | pending | 10 | 7 | review required |
| commit-7037ac15bd8a | [7037ac15bd8a](https://github.com/QuantumNous/new-api/commit/7037ac15bd8a29f8ee3e2b74e784bcdb75d67d22) | pending | 1 | 0 | review required |
| commit-eb48396d5fe9 | [eb48396d5fe9](https://github.com/QuantumNous/new-api/commit/eb48396d5fe97d27772d0cd5e3ca8aa5caa4f3e9) | pending | 336 | 77 | protected files |
| commit-0f2a2075ab07 | [0f2a2075ab07](https://github.com/QuantumNous/new-api/commit/0f2a2075ab072ea7e20ffa5dd5d58dbf1b6b5b22) | pending | 1 | 1 | review required |
| commit-98d50d5383a3 | [98d50d5383a3](https://github.com/QuantumNous/new-api/commit/98d50d5383a33432ff6c30b129461b170e5cbffc) | pending | 1 | 0 | review required |
| commit-b80d633cf586 | [b80d633cf586](https://github.com/QuantumNous/new-api/commit/b80d633cf586b001cfbb4200bae93e65abe57c2b) | pending | 10 | 4 | review required |
| commit-8454082f930f | [8454082f930f](https://github.com/QuantumNous/new-api/commit/8454082f930f44593e92791c2581ffc63eb30a59) | pending | 2 | 0 | review required |
| commit-918427d8ab41 | [918427d8ab41](https://github.com/QuantumNous/new-api/commit/918427d8ab41f6adaa4113d0496f1f8621855b70) | pending | 9 | 7 | review required |
| commit-6c22550ea325 | [6c22550ea325](https://github.com/QuantumNous/new-api/commit/6c22550ea325d4fea0e0ece52412cb1e4449291c) | pending | 23 | 3 | review required |
| commit-66031a09d99f | [66031a09d99f](https://github.com/QuantumNous/new-api/commit/66031a09d99f2ac4e0b94e2c41f04ed691a79304) | pending | 5 | 4 | review required |
| commit-0bee5d441029 | [0bee5d441029](https://github.com/QuantumNous/new-api/commit/0bee5d4410296e972bf0076414ade786c2c799c8) | pending | 2 | 0 | review required |
| commit-dc4732cfed71 | [dc4732cfed71](https://github.com/QuantumNous/new-api/commit/dc4732cfed712b004d3d3d94a414d3ff127a93cc) | pending | 11 | 7 | review required |
| commit-b5b94bc685fd | [b5b94bc685fd](https://github.com/QuantumNous/new-api/commit/b5b94bc685fd2251551df826dab3575ab262f6dc) | pending | 8 | 7 | review required |
| commit-1751f43ee07e | [1751f43ee07e](https://github.com/QuantumNous/new-api/commit/1751f43ee07edc9eb0c56fd9b23586861b43df46) | pending | 1 | 0 | review required |
| commit-6eb6f35ed211 | [6eb6f35ed211](https://github.com/QuantumNous/new-api/commit/6eb6f35ed211b7459cae3b9f13286b9c93fc1bd6) | pending | 5 | 3 | review required |
| commit-b518d0033b67 | [b518d0033b67](https://github.com/QuantumNous/new-api/commit/b518d0033b670f5518b8a2f1cf8ea0142a9d1b8d) | pending | 5 | 4 | protected files |
| commit-74158715cde6 | [74158715cde6](https://github.com/QuantumNous/new-api/commit/74158715cde6d7b767ead23d9a2af64b7b58a588) | pending | 2 | 2 | review required |
| commit-69a41eeadc81 | [69a41eeadc81](https://github.com/QuantumNous/new-api/commit/69a41eeadc81adb08d04346512c76d62fd6203db) | pending | 2 | 1 | review required |
| commit-2bf0820f4b89 | [2bf0820f4b89](https://github.com/QuantumNous/new-api/commit/2bf0820f4b89530acf14d389ba3e8229211933fa) | pending | 2 | 1 | review required |
| commit-2b6f1dfefbe2 | [2b6f1dfefbe2](https://github.com/QuantumNous/new-api/commit/2b6f1dfefbe217fed31fc0726717cc7de6958e8e) | pending | 4 | 2 | protected files |
| commit-8c8c4153d4b8 | [8c8c4153d4b8](https://github.com/QuantumNous/new-api/commit/8c8c4153d4b80d54352d21593de41aa9a6178f7e) | pending | 1 | 1 | previous source reference |
| commit-27ff6a8767e7 | [27ff6a8767e7](https://github.com/QuantumNous/new-api/commit/27ff6a8767e728f879d52770c273d4f73214a430) | pending | 3 | 1 | review required |
| commit-67a0585d0f25 | [67a0585d0f25](https://github.com/QuantumNous/new-api/commit/67a0585d0f252dfca445c11b7600971b7eeb8eea) | pending | 6 | 2 | protected files |
| commit-b7017c251bad | [b7017c251bad](https://github.com/QuantumNous/new-api/commit/b7017c251badaacaab840646a959635d00665e2d) | pending | 2 | 1 | review required |
| commit-0ed497f066a6 | [0ed497f066a6](https://github.com/QuantumNous/new-api/commit/0ed497f066a68613375124303ef54f220267b334) | pending | 149 | 22 | protected files |
| commit-bbd97446c260 | [bbd97446c260](https://github.com/QuantumNous/new-api/commit/bbd97446c26092f2e7250af429096064b9e0f899) | pending | 18 | 2 | review required |
| commit-aece11d2f7f0 | [aece11d2f7f0](https://github.com/QuantumNous/new-api/commit/aece11d2f7f095a33052696c5d28d3656609a99e) | pending | 7 | 2 | review required |
| commit-d8ca0ed0bb59 | [d8ca0ed0bb59](https://github.com/QuantumNous/new-api/commit/d8ca0ed0bb596e910e1955461e7691e40d224d70) | pending | 1 | 1 | protected files |
| commit-73afad588ca7 | [73afad588ca7](https://github.com/QuantumNous/new-api/commit/73afad588ca7af07134fa423e8a33fdea6c855b2) | pending | 2 | 0 | review required |
| commit-057f71c2336c | [057f71c2336c](https://github.com/QuantumNous/new-api/commit/057f71c2336c3981187b732a9d06f65490e9a946) | pending | 25 | 13 | review required |
| commit-219c9e06341f | [219c9e06341f](https://github.com/QuantumNous/new-api/commit/219c9e06341f1b100e2c572a5f97c45f151fd280) | pending | 12 | 2 | review required |
| commit-9f506dd7f905 | [9f506dd7f905](https://github.com/QuantumNous/new-api/commit/9f506dd7f905c288b4a119a8197cd64b77eb4a3f) | pending | 6 | 3 | review required |
| commit-9df450fe54e1 | [9df450fe54e1](https://github.com/QuantumNous/new-api/commit/9df450fe54e1a874a5339b7c38a61014217f02c3) | pending | 36 | 9 | review required |
| commit-36dbbf0f77e7 | [36dbbf0f77e7](https://github.com/QuantumNous/new-api/commit/36dbbf0f77e710455e745048f4a32e8120ad3fd2) | pending | 3 | 1 | review required |
| commit-8f5ab8e4048a | [8f5ab8e4048a](https://github.com/QuantumNous/new-api/commit/8f5ab8e4048a90d88b20ae1e6d5228b04233d3b8) | pending | 1 | 1 | protected files |
| commit-32c261923a97 | [32c261923a97](https://github.com/QuantumNous/new-api/commit/32c261923a9786c64d2af087327ef057e7bde7e3) | pending | 6 | 5 | review required |
| commit-3a9f41ee85cc | [3a9f41ee85cc](https://github.com/QuantumNous/new-api/commit/3a9f41ee85cc369f5b8d7fe6e62ff4e7bf3a9ec8) | pending | 2 | 2 | review required |
| commit-7c044d7c5c2d | [7c044d7c5c2d](https://github.com/QuantumNous/new-api/commit/7c044d7c5c2d2beadf16b21910950f8f593bc3ef) | pending | 46 | 18 | previous source reference |
| commit-6b659fd61c50 | [6b659fd61c50](https://github.com/QuantumNous/new-api/commit/6b659fd61c50e35d559c41520a0fff7b8aea56a4) | pending | 11 | 2 | review required |
| commit-d5803532bdcc | [d5803532bdcc](https://github.com/QuantumNous/new-api/commit/d5803532bdccde3a2b1583291f51e92d3519b1c6) | pending | 1 | 1 | protected files |
| commit-eb99ab1b4034 | [eb99ab1b4034](https://github.com/QuantumNous/new-api/commit/eb99ab1b40343c3317bb47981cccdbb2b159a5fa) | pending | 5 | 1 | review required |
| commit-2cf177ac487e | [2cf177ac487e](https://github.com/QuantumNous/new-api/commit/2cf177ac487e62c627c7d423b65735ba2481ef4f) | pending | 2 | 0 | review required |
| commit-49ec46966825 | [49ec46966825](https://github.com/QuantumNous/new-api/commit/49ec4696682530781a036eab1ac195f0b04706c0) | pending | 4 | 3 | review required |
| commit-d8cb177440ce | [d8cb177440ce](https://github.com/QuantumNous/new-api/commit/d8cb177440ceaae422d5bfd96c258d47af4e0f1d) | pending | 129 | 42 | review required |
| commit-9a8674425c5a | [9a8674425c5a](https://github.com/QuantumNous/new-api/commit/9a8674425c5a43435a259b58bb928a55d26be990) | pending | 7 | 4 | protected files |
| commit-45c3fbe8aeb0 | [45c3fbe8aeb0](https://github.com/QuantumNous/new-api/commit/45c3fbe8aeb049f03c13e14298a40b87aea5bd87) | pending | 74 | 24 | protected files |
| commit-3e84ec0ab823 | [3e84ec0ab823](https://github.com/QuantumNous/new-api/commit/3e84ec0ab8239cf2277f8a10d45566630a9c10fa) | pending | 49 | 24 | review required |
| commit-a8729b5c3709 | [a8729b5c3709](https://github.com/QuantumNous/new-api/commit/a8729b5c3709cc01d88fc3f2db5b91347fc9129e) | pending | 12 | 3 | review required |
| commit-0973dc2b8f55 | [0973dc2b8f55](https://github.com/QuantumNous/new-api/commit/0973dc2b8f550de71b75fdd3805576d3ce6ccf42) | pending | 61 | 24 | review required |
| commit-3f8a50cf8877 | [3f8a50cf8877](https://github.com/QuantumNous/new-api/commit/3f8a50cf8877683669cd812240a0beaf7b171c32) | pending | 33 | 18 | protected files |
| commit-521cebf585ef | [521cebf585ef](https://github.com/QuantumNous/new-api/commit/521cebf585efc2e782dd9fb93d0f66752c8d3c32) | pending | 3 | 1 | review required |
| commit-6f2333990613 | [6f2333990613](https://github.com/QuantumNous/new-api/commit/6f2333990613bf3e9dd36f541fc380148c7b5175) | pending | 55 | 24 | review required |
| commit-0c76e4dae77a | [0c76e4dae77a](https://github.com/QuantumNous/new-api/commit/0c76e4dae77a279e015329b7478e6f02d6b62edd) | pending | 122 | 19 | review required |
| commit-3b4652269a6a | [3b4652269a6a](https://github.com/QuantumNous/new-api/commit/3b4652269a6a6d9e2c8650a84ee8c4e447e6473b) | pending | 2 | 0 | review required |
| commit-6e10f9bc927a | [6e10f9bc927a](https://github.com/QuantumNous/new-api/commit/6e10f9bc927a4eae889864a6ef601359d53526b9) | pending | 7 | 6 | previous source reference |
| commit-7bbe85bcb095 | [7bbe85bcb095](https://github.com/QuantumNous/new-api/commit/7bbe85bcb09546e0b89572bf97198fc94889be3d) | pending | 5 | 1 | review required |
| commit-387a40914853 | [387a40914853](https://github.com/QuantumNous/new-api/commit/387a40914853310d69adc2f52474134ced5f4811) | pending | 6 | 0 | review required |
| commit-5c7cca015525 | [5c7cca015525](https://github.com/QuantumNous/new-api/commit/5c7cca015525212a7ac2741da7f3b51fa2e30db0) | pending | 5 | 0 | review required |
| commit-6298b0f32384 | [6298b0f32384](https://github.com/QuantumNous/new-api/commit/6298b0f3238461b9629dfc1c00866f8325123aa1) | pending | 2 | 0 | review required |
| commit-92bc7ff73c5e | [92bc7ff73c5e](https://github.com/QuantumNous/new-api/commit/92bc7ff73c5ef215496d09d3e4b5769c9e35006b) | pending | 3 | 0 | review required |
| commit-210734bb73bc | [210734bb73bc](https://github.com/QuantumNous/new-api/commit/210734bb73bc3c6e37548af90360aa4226566fe2) | pending | 12 | 2 | review required |
| commit-bee45b58a3c0 | [bee45b58a3c0](https://github.com/QuantumNous/new-api/commit/bee45b58a3c0b77e8dc81e6b5aeb4474aa9058d1) | pending | 3 | 0 | review required |
| commit-75e533209490 | [75e533209490](https://github.com/QuantumNous/new-api/commit/75e533209490a8ef3a8b5e3d93e4dac03ba19bcf) | pending | 18 | 7 | review required |
| commit-99974a814f00 | [99974a814f00](https://github.com/QuantumNous/new-api/commit/99974a814f00a26dfa7431beb313e487529bbeed) | pending | 29 | 2 | review required |
| commit-984330920061 | [984330920061](https://github.com/QuantumNous/new-api/commit/984330920061e014159c7581b8fcdb726de2c247) | pending | 43 | 7 | review required |
| commit-eb76b136b85e | [eb76b136b85e](https://github.com/QuantumNous/new-api/commit/eb76b136b85ecba9d6ad19c714c5781e51815530) | pending | 12 | 4 | review required |
| commit-0e0ba152bdcc | [0e0ba152bdcc](https://github.com/QuantumNous/new-api/commit/0e0ba152bdcc6891f6053047ccf14d41b3cad60a) | pending | 29 | 2 | review required |
| commit-71c1fd7caad7 | [71c1fd7caad7](https://github.com/QuantumNous/new-api/commit/71c1fd7caad738db4d13aabbf28eeadb293d0cfe) | pending | 36 | 11 | review required |
| commit-a5e41a893379 | [a5e41a893379](https://github.com/QuantumNous/new-api/commit/a5e41a893379e49bd9e1d025e775c04df8961c95) | pending | 20 | 11 | review required |
| commit-2bec370629aa | [2bec370629aa](https://github.com/QuantumNous/new-api/commit/2bec370629aa72d74d3f598b953a4d33b8bcbf7d) | pending | 28 | 9 | review required |
| commit-8f72ecbbfd86 | [8f72ecbbfd86](https://github.com/QuantumNous/new-api/commit/8f72ecbbfd86ded5ee15373921533eed3d334a9a) | pending | 5 | 1 | review required |
| commit-551bb63edf40 | [551bb63edf40](https://github.com/QuantumNous/new-api/commit/551bb63edf4007d7c4b0930504faf6db87012d73) | pending | 2 | 0 | review required |
| commit-bd22e45a740a | [bd22e45a740a](https://github.com/QuantumNous/new-api/commit/bd22e45a740a7c02c328704dc419d8795237cb74) | pending | 1 | 0 | review required |
| commit-950644c9d544 | [950644c9d544](https://github.com/QuantumNous/new-api/commit/950644c9d54445bdd8796643e4e2e18e13167a4c) | pending | 3 | 0 | review required |
| commit-524455fac3c4 | [524455fac3c4](https://github.com/QuantumNous/new-api/commit/524455fac3c438321df4ae9ed4ffd11bc635cc4f) | pending | 23 | 10 | review required |
| commit-fff0635bb14b | [fff0635bb14b](https://github.com/QuantumNous/new-api/commit/fff0635bb14b8ec5f10df9e582a83591df841902) | pending | 1 | 1 | protected files |
| commit-ebe4c368f287 | [ebe4c368f287](https://github.com/QuantumNous/new-api/commit/ebe4c368f28787919ea478c0abf1ea7101978179) | pending | 214 | 46 | review required |
| commit-ea7cb0ba4e0f | [ea7cb0ba4e0f](https://github.com/QuantumNous/new-api/commit/ea7cb0ba4e0f82e2bfa5e55752eb68bdf902f71b) | pending | 17 | 1 | review required |
| commit-9bf328d97497 | [9bf328d97497](https://github.com/QuantumNous/new-api/commit/9bf328d9749751757d5d6b74088d514813a618bd) | pending | 5 | 0 | protected files |
| commit-4fc9d1f1fa77 | [4fc9d1f1fa77](https://github.com/QuantumNous/new-api/commit/4fc9d1f1fa77c0cfdd9719cb59ff9ecc9885d66a) | pending | 3 | 1 | review required |
| commit-c79b74b68358 | [c79b74b68358](https://github.com/QuantumNous/new-api/commit/c79b74b68358180c68440057596bdc34a99cc649) | pending | 11 | 3 | review required |
| commit-876903a8eb22 | [876903a8eb22](https://github.com/QuantumNous/new-api/commit/876903a8eb22c44e395c03da38f6701c650651ae) | pending | 1 | 0 | review required |
| commit-d52bdc0b4087 | [d52bdc0b4087](https://github.com/QuantumNous/new-api/commit/d52bdc0b4087d50d87eb06387e86b2be53e85cd4) | pending | 48 | 8 | protected files |
| commit-12be9975c0bf | [12be9975c0bf](https://github.com/QuantumNous/new-api/commit/12be9975c0bf01fa175a2bb3360607767c8ba5fb) | pending | 174 | 26 | protected files |
| commit-a20574136b27 | [a20574136b27](https://github.com/QuantumNous/new-api/commit/a20574136b2746e9afc4b268c99cf99b3f6cc68b) | pending | 3 | 0 | review required |
| commit-7cf9b473f61e | [7cf9b473f61e](https://github.com/QuantumNous/new-api/commit/7cf9b473f61eb7dc09f6f0782192f17595140a08) | pending | 2 | 0 | review required |
| commit-064ed943e1ac | [064ed943e1ac](https://github.com/QuantumNous/new-api/commit/064ed943e1ac40e3eaca1b58ffb7fa5dacb3fde3) | pending | 54 | 14 | protected files |
| commit-bdef11750524 | [bdef11750524](https://github.com/QuantumNous/new-api/commit/bdef117505247769268b209665fb3ad7554c3da7) | pending | 1 | 0 | review required |
| commit-b6566f33d908 | [b6566f33d908](https://github.com/QuantumNous/new-api/commit/b6566f33d908fee574c8a6da2a23c4fd5df7353c) | pending | 10 | 8 | review required |
| commit-3e1b8b153249 | [3e1b8b153249](https://github.com/QuantumNous/new-api/commit/3e1b8b15324941ff20c77616737d8dd984582d33) | pending | 16 | 7 | review required |
| commit-3cea2bf799a6 | [3cea2bf799a6](https://github.com/QuantumNous/new-api/commit/3cea2bf799a61e07ad53368538fdb7064b017d77) | pending | 4 | 0 | review required |
| commit-f362c7c51b86 | [f362c7c51b86](https://github.com/QuantumNous/new-api/commit/f362c7c51b86a97e1fc72e4d48280cf53a5d2cb9) | pending | 8 | 0 | review required |
| commit-f064bffa2b65 | [f064bffa2b65](https://github.com/QuantumNous/new-api/commit/f064bffa2b65dd2c6348b375f8b91bab41aabc19) | pending | 25 | 4 | review required |
| commit-f256e40bcb29 | [f256e40bcb29](https://github.com/QuantumNous/new-api/commit/f256e40bcb29292376edb5f0ba2e74ffa463aced) | pending | 3 | 0 | review required |
| commit-25ec832faa48 | [25ec832faa48](https://github.com/QuantumNous/new-api/commit/25ec832faa48031dd74f6fff0f58390e29a79f7c) | pending | 8 | 2 | review required |
| commit-39294418af4c | [39294418af4c](https://github.com/QuantumNous/new-api/commit/39294418af4c920950be1b8c1fed0ed0d2bea52e) | pending | 37 | 12 | protected files |
| commit-b3e279464c82 | [b3e279464c82](https://github.com/QuantumNous/new-api/commit/b3e279464c82d6b1bd9fca6ce27f67c4c9246fec) | pending | 10 | 0 | review required |
| commit-505805a4c3ab | [505805a4c3ab](https://github.com/QuantumNous/new-api/commit/505805a4c3ab55fbcf6227ec110226ff1c193e0b) | pending | 37 | 12 | review required |
| commit-d4c26bfb8dcb | [d4c26bfb8dcb](https://github.com/QuantumNous/new-api/commit/d4c26bfb8dcb7434a0bc1e0d6153e848d86c39b4) | pending | 3 | 2 | review required |
| commit-ab489ab8842b | [ab489ab8842b](https://github.com/QuantumNous/new-api/commit/ab489ab8842b26f3dbc2d7a384e7c0ab0f700e0e) | pending | 19 | 11 | review required |
| commit-251b76d8633e | [251b76d8633e](https://github.com/QuantumNous/new-api/commit/251b76d8633ec4eeb1064dcdb5eb806ecf1a8668) | pending | 19 | 9 | review required |
| commit-74629e29f83f | [74629e29f83f](https://github.com/QuantumNous/new-api/commit/74629e29f83f506bb523c5542aa851947bd423eb) | pending | 96 | 24 | protected files |
| commit-385d2dfd10d8 | [385d2dfd10d8](https://github.com/QuantumNous/new-api/commit/385d2dfd10d821b25c8a6766bd16eea248cb1652) | pending | 48 | 26 | review required |
| commit-129f21b6942e | [129f21b6942e](https://github.com/QuantumNous/new-api/commit/129f21b6942ed433fd7eea1239b4a24d8b58b48b) | pending | 2 | 0 | review required |
| commit-c9a110190c52 | [c9a110190c52](https://github.com/QuantumNous/new-api/commit/c9a110190c5241c24d9d66de340431f5e9873db6) | pending | 10 | 7 | review required |
| commit-be36cbb8facf | [be36cbb8facf](https://github.com/QuantumNous/new-api/commit/be36cbb8facf99d523197f9e18d02b7605331048) | pending | 2 | 0 | review required |
| commit-007d69942d24 | [007d69942d24](https://github.com/QuantumNous/new-api/commit/007d69942d24eeea9edd4665f812d07ed11b3b97) | pending | 2 | 0 | review required |
| commit-2ba61576146f | [2ba61576146f](https://github.com/QuantumNous/new-api/commit/2ba61576146f0583f789ee6845f2280cb86819ce) | pending | 6 | 1 | review required |
| commit-043ff99a51ec | [043ff99a51ec](https://github.com/QuantumNous/new-api/commit/043ff99a51ecad8229389ddd04f45f4b25a23ac6) | pending | 5 | 1 | review required |
| commit-33142f0aee7c | [33142f0aee7c](https://github.com/QuantumNous/new-api/commit/33142f0aee7c87b2110f25dbddcb861dbdc2c44e) | pending | 4 | 1 | review required |
| commit-7fd06381976c | [7fd06381976c](https://github.com/QuantumNous/new-api/commit/7fd06381976cbea126d1f20973548bf60f20e37d) | pending | 3 | 1 | review required |
| commit-815217ba648f | [815217ba648f](https://github.com/QuantumNous/new-api/commit/815217ba648ff800c4f583fbfa36ba286cee6b2d) | pending | 2 | 0 | review required |
| commit-d92612038410 | [d92612038410](https://github.com/QuantumNous/new-api/commit/d926120384109a9e55b0cd54a06bb6e3ae860a2c) | pending | 7 | 0 | review required |
| commit-d1c79d728822 | [d1c79d728822](https://github.com/QuantumNous/new-api/commit/d1c79d7288221768a5f7270649d691ac5a137def) | pending | 1 | 0 | review required |
| commit-2509e25fa060 | [2509e25fa060](https://github.com/QuantumNous/new-api/commit/2509e25fa0601d26f5b8fbb3194ec76689a91a77) | pending | 3 | 0 | review required |
| commit-76f7dafd2b82 | [76f7dafd2b82](https://github.com/QuantumNous/new-api/commit/76f7dafd2b82fc305533ce6a3065b3de2d338bae) | pending | 2 | 0 | review required |
| commit-04c64734cc7c | [04c64734cc7c](https://github.com/QuantumNous/new-api/commit/04c64734cc7c2e58a9efd0b247182330f3f52cce) | pending | 2 | 0 | review required |
| commit-9fe0457ee1f4 | [9fe0457ee1f4](https://github.com/QuantumNous/new-api/commit/9fe0457ee1f4b9de407a254500d54f5a8f41ee29) | pending | 52 | 20 | review required |
| commit-8529f209c859 | [8529f209c859](https://github.com/QuantumNous/new-api/commit/8529f209c85913de9ec98c5af49dcfce4a41361a) | pending | 59 | 29 | review required |
| commit-2bfb89c1b99d | [2bfb89c1b99d](https://github.com/QuantumNous/new-api/commit/2bfb89c1b99d5eb7a799ba5706813614bfb6189a) | pending | 1 | 0 | review required |
| commit-8e5e09166cf8 | [8e5e09166cf8](https://github.com/QuantumNous/new-api/commit/8e5e09166cf83731df1819fd1cb9c50ea2576ed6) | pending | 2 | 0 | review required |
| commit-62f8db775bcd | [62f8db775bcd](https://github.com/QuantumNous/new-api/commit/62f8db775bcdaa5da144d9c4a36f5ead82de3ac5) | pending | 2 | 0 | review required |
| commit-d3874db61fce | [d3874db61fce](https://github.com/QuantumNous/new-api/commit/d3874db61fce61b6c0f5c37523b21912a3f2658d) | pending | 2 | 0 | review required |
| commit-5caafd3d84dc | [5caafd3d84dc](https://github.com/QuantumNous/new-api/commit/5caafd3d84dc74c8b0d081524f57a534b3980cb0) | pending | 16 | 13 | review required |
| commit-fa90b2312cfb | [fa90b2312cfb](https://github.com/QuantumNous/new-api/commit/fa90b2312cfb686aaa6010f9280bffcf0bf6d391) | pending | 6 | 1 | review required |
| commit-81336fc69b8a | [81336fc69b8a](https://github.com/QuantumNous/new-api/commit/81336fc69b8a932e70c24a8dbf439bec94c98328) | pending | 1 | 1 | review required |
| commit-610334dbd760 | [610334dbd760](https://github.com/QuantumNous/new-api/commit/610334dbd760f1c5199a09ae5f860f11451f4c56) | pending | 3 | 1 | review required |
| commit-69a50029819a | [69a50029819a](https://github.com/QuantumNous/new-api/commit/69a50029819a26c53e6babd276d49cfe2f8880ad) | pending | 18 | 13 | review required |
| commit-7209b6db95ef | [7209b6db95ef](https://github.com/QuantumNous/new-api/commit/7209b6db95ef33bfb0f6344ebf61c1fec9af780c) | pending | 5 | 4 | review required |
| commit-57791e31c62b | [57791e31c62b](https://github.com/QuantumNous/new-api/commit/57791e31c62bdf97df855f3b7dc912046b3bc990) | pending | 9 | 2 | protected files |
| commit-ed0ad6e3ff9f | [ed0ad6e3ff9f](https://github.com/QuantumNous/new-api/commit/ed0ad6e3ff9f8785a19d93cba73cf06ef2219668) | pending | 11 | 8 | review required |
| commit-07578faf26d1 | [07578faf26d1](https://github.com/QuantumNous/new-api/commit/07578faf26d1a21c70bf0fbf29f12730a212c583) | pending | 2 | 0 | review required |
| commit-d82678a9708b | [d82678a9708b](https://github.com/QuantumNous/new-api/commit/d82678a9708b611d52386c01f04238ec01f07463) | pending | 3 | 1 | protected files |
| commit-0cde9d94f6d3 | [0cde9d94f6d3](https://github.com/QuantumNous/new-api/commit/0cde9d94f6d3b77d5258ca6b36959f895d26f2a4) | pending | 31 | 8 | review required |
| commit-8b2c7105f91f | [8b2c7105f91f](https://github.com/QuantumNous/new-api/commit/8b2c7105f91f2316f90eb86f9b1c6b98774c7105) | pending | 1 | 0 | review required |
| commit-73a471f3a8b1 | [73a471f3a8b1](https://github.com/QuantumNous/new-api/commit/73a471f3a8b1cdf5dfe371184dfc3579cdcbefcc) | pending | 71 | 23 | review required |
| commit-8a11b4305ed1 | [8a11b4305ed1](https://github.com/QuantumNous/new-api/commit/8a11b4305ed139b55ee91f914c6a19d59c783d72) | pending | 25 | 1 | review required |
| commit-a5b40663d330 | [a5b40663d330](https://github.com/QuantumNous/new-api/commit/a5b40663d3307fc5011c85b83eb18bbaa33dbad2) | pending | 40 | 13 | review required |
| commit-6237d9d77eb0 | [6237d9d77eb0](https://github.com/QuantumNous/new-api/commit/6237d9d77eb0544329b88ff96d71c1715af4268f) | pending | 3 | 0 | review required |
| commit-75f3d246e2b1 | [75f3d246e2b1](https://github.com/QuantumNous/new-api/commit/75f3d246e2b1d6a3ea56342d9bc82cb3e1d0b6e2) | pending | 3 | 0 | review required |
| commit-90134d5b6bff | [90134d5b6bff](https://github.com/QuantumNous/new-api/commit/90134d5b6bff1a66affa35c5bd8e90e3d74a93f8) | pending | 21 | 15 | review required |
| commit-a8ed7f7c5f24 | [a8ed7f7c5f24](https://github.com/QuantumNous/new-api/commit/a8ed7f7c5f24e56c2f1612126d82e9fec625056c) | pending | 2 | 0 | review required |
| commit-e2471403c779 | [e2471403c779](https://github.com/QuantumNous/new-api/commit/e2471403c77957818b939ad86bbeccd051d2d042) | pending | 2 | 0 | review required |
| commit-42cbe6206a62 | [42cbe6206a62](https://github.com/QuantumNous/new-api/commit/42cbe6206a62550896c8371d62a6f15bad5b3e05) | pending | 16 | 8 | protected files |
| commit-d8069436612e | [d8069436612e](https://github.com/QuantumNous/new-api/commit/d8069436612e08bd0d96169222b10031556d09a7) | pending | 9 | 7 | review required |
| commit-ae249f4ecbaf | [ae249f4ecbaf](https://github.com/QuantumNous/new-api/commit/ae249f4ecbaf3ae9b1fc62cf377ca5a764b9a002) | pending | 11 | 3 | review required |
| commit-fa3cc1c6b9b1 | [fa3cc1c6b9b1](https://github.com/QuantumNous/new-api/commit/fa3cc1c6b9b13f7bf25c3e896d439185d326ac9c) | pending | 14 | 8 | review required |
| commit-32ef6d216f9f | [32ef6d216f9f](https://github.com/QuantumNous/new-api/commit/32ef6d216f9fa0a2aa182bd8aad9a02daf3767c8) | pending | 14 | 8 | review required |
| commit-9e905a83e1d6 | [9e905a83e1d6](https://github.com/QuantumNous/new-api/commit/9e905a83e1d6c369b3aa43de4f04ab244cce3205) | pending | 5 | 2 | review required |
| commit-ed7c4e35d429 | [ed7c4e35d429](https://github.com/QuantumNous/new-api/commit/ed7c4e35d4297962991ac5f00eea504ceac0fbd0) | pending | 32 | 14 | review required |
| commit-e2bb05f70ed6 | [e2bb05f70ed6](https://github.com/QuantumNous/new-api/commit/e2bb05f70ed6b6725a83ff6f1c97dbd2c2f640d4) | pending | 2 | 0 | review required |
| commit-55a6cd2a44ea | [55a6cd2a44ea](https://github.com/QuantumNous/new-api/commit/55a6cd2a44eaf189bbd565dd2cdfa591d19b06d6) | pending | 2 | 0 | review required |
| commit-3524fe0b1579 | [3524fe0b1579](https://github.com/QuantumNous/new-api/commit/3524fe0b15794d8d19378827d36a7edc0b0e91ea) | pending | 4 | 0 | protected files |
| commit-2906e4f779b7 | [2906e4f779b7](https://github.com/QuantumNous/new-api/commit/2906e4f779b715f282ae11203211dca77051d5af) | pending | 18 | 7 | review required |
| commit-3e8c358da09b | [3e8c358da09b](https://github.com/QuantumNous/new-api/commit/3e8c358da09b4598a946712db8219ff44f197475) | pending | 1 | 0 | review required |
| commit-946c730ed008 | [946c730ed008](https://github.com/QuantumNous/new-api/commit/946c730ed00806650fef509c74a868dda2847f7c) | pending | 3 | 0 | review required |
| commit-ea336ed6654a | [ea336ed6654a](https://github.com/QuantumNous/new-api/commit/ea336ed6654a97c0c27e004581578231e0d92824) | pending | 7 | 2 | review required |
| commit-96b0da2273a1 | [96b0da2273a1](https://github.com/QuantumNous/new-api/commit/96b0da2273a187b3fd9621f77b10dcebd046e2f6) | pending | 6 | 1 | review required |
| commit-1ea24a050d5e | [1ea24a050d5e](https://github.com/QuantumNous/new-api/commit/1ea24a050d5e276584d744122810b8dc57b10f0e) | pending | 6 | 2 | protected files |
| commit-e04c05c2b929 | [e04c05c2b929](https://github.com/QuantumNous/new-api/commit/e04c05c2b929c0df6cc35227f860cf6e19d480c7) | pending | 8 | 1 | review required |
| commit-b0bf2580e249 | [b0bf2580e249](https://github.com/QuantumNous/new-api/commit/b0bf2580e2491869ee9fe861c12b86153c3fe5f4) | pending | 6 | 2 | protected files |
| commit-972aed197282 | [972aed197282](https://github.com/QuantumNous/new-api/commit/972aed1972820389ea0b603ca58f03f846fbf790) | pending | 11 | 1 | review required |
| commit-a89037d0b9b0 | [a89037d0b9b0](https://github.com/QuantumNous/new-api/commit/a89037d0b9b065da8582959018321abbc48db549) | pending | 3 | 1 | review required |
| commit-ef70d1b79998 | [ef70d1b79998](https://github.com/QuantumNous/new-api/commit/ef70d1b79998db9bb32b6b79205afbc4b9f8716b) | pending | 1 | 1 | review required |
| commit-b932e6d876cd | [b932e6d876cd](https://github.com/QuantumNous/new-api/commit/b932e6d876cda4fa49528b5cdb873d9fc5e31b6f) | pending | 21 | 8 | review required |
| commit-9c3d3aeb3c30 | [9c3d3aeb3c30](https://github.com/QuantumNous/new-api/commit/9c3d3aeb3c30ea09c7ad0f810c86205c4da558d7) | pending | 16 | 3 | review required |
| commit-33320407f4f5 | [33320407f4f5](https://github.com/QuantumNous/new-api/commit/33320407f4f5dcf5bc635ee6d73dcc31c8bcc031) | pending | 8 | 6 | review required |
| commit-4c34f25a49f1 | [4c34f25a49f1](https://github.com/QuantumNous/new-api/commit/4c34f25a49f1cf491479e3842b2251e8a6908851) | pending | 48 | 19 | review required |
| commit-2d7aef7414e8 | [2d7aef7414e8](https://github.com/QuantumNous/new-api/commit/2d7aef7414e8671b1ac7db336bc170eeb7cc5e2b) | pending | 2 | 0 | review required |
| commit-6b638788c645 | [6b638788c645](https://github.com/QuantumNous/new-api/commit/6b638788c645a6c9164c5c6394827e3f5e3a1abd) | pending | 12 | 7 | review required |
| commit-db536ec8bcbb | [db536ec8bcbb](https://github.com/QuantumNous/new-api/commit/db536ec8bcbbc63034d009a7257ecee0f3468abf) | pending | 17 | 8 | review required |
| commit-58a9eff4c842 | [58a9eff4c842](https://github.com/QuantumNous/new-api/commit/58a9eff4c8421a6bbb04855e561310d3eeabd107) | pending | 8 | 4 | review required |
| commit-3fdf9083dc6b | [3fdf9083dc6b](https://github.com/QuantumNous/new-api/commit/3fdf9083dc6bfc47fc2ad0de3fac327d749d778d) | pending | 7 | 5 | review required |
| commit-da2540dda132 | [da2540dda132](https://github.com/QuantumNous/new-api/commit/da2540dda1328040193f279a58c3e315083e416c) | pending | 1 | 0 | review required |
| commit-d904b9700015 | [d904b9700015](https://github.com/QuantumNous/new-api/commit/d904b9700015ad73285b0f0030c88eb6683fbeda) | pending | 13 | 7 | review required |
| commit-585c4610e5d7 | [585c4610e5d7](https://github.com/QuantumNous/new-api/commit/585c4610e5d7103480620ab61954db5a5feac35b) | pending | 9 | 7 | review required |
| commit-c49972441e73 | [c49972441e73](https://github.com/QuantumNous/new-api/commit/c49972441e73f05207427567a2d9ad7c620db4bd) | pending | 23 | 3 | review required |
| commit-03563a4a7916 | [03563a4a7916](https://github.com/QuantumNous/new-api/commit/03563a4a791629a811076f1cbcbecfbe2716eac2) | pending | 80 | 22 | protected files |
| commit-a5bf94022060 | [a5bf94022060](https://github.com/QuantumNous/new-api/commit/a5bf940220604e5b9d150e2338bf6528dfbb8be4) | pending | 9 | 0 | review required |
| commit-2fddb863f221 | [2fddb863f221](https://github.com/QuantumNous/new-api/commit/2fddb863f2219d9eacea5b899ee39b46f11adf33) | pending | 9 | 7 | review required |
| commit-3abbb8198e04 | [3abbb8198e04](https://github.com/QuantumNous/new-api/commit/3abbb8198e0468ac6e16eb0103c1d3642e2e7668) | pending | 3 | 1 | review required |
| commit-65d3a2171fd1 | [65d3a2171fd1](https://github.com/QuantumNous/new-api/commit/65d3a2171fd11c65ad2e2e1f1d65723926659c81) | pending | 4 | 0 | review required |
| commit-9978ee1e25a6 | [9978ee1e25a6](https://github.com/QuantumNous/new-api/commit/9978ee1e25a647bfe004e96c8719a2cb62c24732) | pending | 23 | 12 | review required |
| commit-47713bcb1749 | [47713bcb1749](https://github.com/QuantumNous/new-api/commit/47713bcb1749f81c3a48127c0150e4c6d4da35f1) | pending | 4 | 0 | review required |
| commit-9a0be8750a6d | [9a0be8750a6d](https://github.com/QuantumNous/new-api/commit/9a0be8750a6d736d9692535ed2cd68f8eec46529) | pending | 7 | 0 | protected files |
| commit-9c293e8c0237 | [9c293e8c0237](https://github.com/QuantumNous/new-api/commit/9c293e8c02371bda844af79e3500ff2d516d1dda) | pending | 1 | 0 | review required |
| commit-474ed66fb14b | [474ed66fb14b](https://github.com/QuantumNous/new-api/commit/474ed66fb14b0de30f82a1729b4f7cf3dab19d63) | pending | 2 | 0 | review required |
| commit-00e8a00cb4ae | [00e8a00cb4ae](https://github.com/QuantumNous/new-api/commit/00e8a00cb4aebb4e029b5ff32b8fe1e98f0b0bc5) | pending | 2 | 0 | review required |
| commit-e537dc380bb2 | [e537dc380bb2](https://github.com/QuantumNous/new-api/commit/e537dc380bb253d4f902ac033ab39dd2ca462b5e) | pending | 9 | 7 | review required |
| commit-c0cff23a3af8 | [c0cff23a3af8](https://github.com/QuantumNous/new-api/commit/c0cff23a3af8e68809647bd39ec204bedbea28af) | pending | 17 | 7 | review required |
| commit-2c175190c44a | [2c175190c44a](https://github.com/QuantumNous/new-api/commit/2c175190c44a95c6af4877e901dc5211ef706979) | pending | 3 | 0 | review required |
| commit-d61d6be7593f | [d61d6be7593f](https://github.com/QuantumNous/new-api/commit/d61d6be7593f99eacf9462e99ee13f0c432657f5) | pending | 3 | 1 | review required |
| commit-54eee488bed3 | [54eee488bed3](https://github.com/QuantumNous/new-api/commit/54eee488bed3ec6b9659b3a792194b18c2dfad27) | pending | 6 | 1 | review required |
| commit-6e9de44a7dca | [6e9de44a7dca](https://github.com/QuantumNous/new-api/commit/6e9de44a7dca08f44234ab9c623f6a93a174d0c4) | pending | 2 | 0 | review required |
| commit-4eb3b9160566 | [4eb3b9160566](https://github.com/QuantumNous/new-api/commit/4eb3b9160566dc4c1f4f0e9340de254f7febf056) | pending | 5 | 0 | review required |
| commit-b6809a52df88 | [b6809a52df88](https://github.com/QuantumNous/new-api/commit/b6809a52df88e8db5c8591b2d3788a05401271b7) | pending | 3 | 0 | review required |
| commit-0aec08fee811 | [0aec08fee811](https://github.com/QuantumNous/new-api/commit/0aec08fee811ec6136828fda790551b49e410301) | pending | 0 | 0 | review required |
| commit-9310231b3c27 | [9310231b3c27](https://github.com/QuantumNous/new-api/commit/9310231b3c27fea933e939b46cf26e0ce67192e3) | pending | 2 | 0 | review required |
| commit-8cb88ebd17a0 | [8cb88ebd17a0](https://github.com/QuantumNous/new-api/commit/8cb88ebd17a072691f39705ae72c786f8bb5c4e0) | pending | 2 | 0 | review required |
| commit-996adffe5165 | [996adffe5165](https://github.com/QuantumNous/new-api/commit/996adffe5165bd5e311e33a03a86b8aede1fe376) | pending | 2 | 0 | review required |
| commit-5401874c6387 | [5401874c6387](https://github.com/QuantumNous/new-api/commit/5401874c6387b52149b7e5b0f61d9dafea5f5b7a) | pending | 2 | 2 | protected files |
| commit-a46f045d6d14 | [a46f045d6d14](https://github.com/QuantumNous/new-api/commit/a46f045d6d146b1e7e17616974bbf37e32885850) | pending | 1 | 0 | review required |
| commit-6c14c0762dc7 | [6c14c0762dc7](https://github.com/QuantumNous/new-api/commit/6c14c0762dc7f3bc77d0c80585f974e65ac2a571) | pending | 1 | 0 | review required |
| commit-c76452d22412 | [c76452d22412](https://github.com/QuantumNous/new-api/commit/c76452d22412be45eb708f2618748f7aac46703d) | pending | 1 | 0 | review required |
| commit-d04c118c8803 | [d04c118c8803](https://github.com/QuantumNous/new-api/commit/d04c118c8803f49e0c9bab74dcf5b5efeab9464a) | pending | 2 | 1 | review required |
| commit-c2b7a9a9e0b5 | [c2b7a9a9e0b5](https://github.com/QuantumNous/new-api/commit/c2b7a9a9e0b548c2051a949fceabb59029adcb49) | pending | 2 | 0 | review required |
| commit-789c970199ea | [789c970199ea](https://github.com/QuantumNous/new-api/commit/789c970199ea527e6a26e071915f4a4cd2c64178) | pending | 1 | 0 | review required |
| commit-f5e54ccdb203 | [f5e54ccdb203](https://github.com/QuantumNous/new-api/commit/f5e54ccdb203e8d4469035e72cc48dfd25b4314d) | pending | 7 | 0 | review required |
| commit-87bb71e7fc83 | [87bb71e7fc83](https://github.com/QuantumNous/new-api/commit/87bb71e7fc83121fc995a048077e5c0e209b789f) | pending | 2 | 0 | review required |
| commit-3509ae7c194d | [3509ae7c194d](https://github.com/QuantumNous/new-api/commit/3509ae7c194dcc0011cc4ec20cd495b198a7ed6d) | pending | 6 | 0 | review required |
| commit-6cbb1c7edba4 | [6cbb1c7edba4](https://github.com/QuantumNous/new-api/commit/6cbb1c7edba4d634ba637a6e6f111115d6de9c9a) | pending | 29 | 16 | review required |
| commit-caca52f8dea0 | [caca52f8dea0](https://github.com/QuantumNous/new-api/commit/caca52f8dea0436d63bacc0dd3e0a73ba7b6b094) | pending | 79 | 32 | review required |
| commit-335ebcff71b8 | [335ebcff71b8](https://github.com/QuantumNous/new-api/commit/335ebcff71b843655bf2c64f7a9a33ca21b792ea) | pending | 16 | 1 | review required |
| commit-2ffc59590115 | [2ffc59590115](https://github.com/QuantumNous/new-api/commit/2ffc59590115bb99d5252ceaeb3e73f06dfcd1ae) | pending | 24 | 4 | protected files |
| commit-c71eefbcd194 | [c71eefbcd194](https://github.com/QuantumNous/new-api/commit/c71eefbcd194ea3caca8476e399ae580b275470c) | pending | 14 | 2 | protected files |
| commit-2506e1b980ec | [2506e1b980ec](https://github.com/QuantumNous/new-api/commit/2506e1b980ecf67c829ef20255a16fc61c47cb55) | pending | 2 | 1 | review required |
| commit-4924361aef73 | [4924361aef73](https://github.com/QuantumNous/new-api/commit/4924361aef733ed724cfdb77ccf35690d9629971) | pending | 34 | 13 | review required |
| commit-ae73ef8e25b9 | [ae73ef8e25b9](https://github.com/QuantumNous/new-api/commit/ae73ef8e25b957ea3291512b66664f418f33c180) | pending | 2 | 0 | review required |
| commit-2035a82aeb54 | [2035a82aeb54](https://github.com/QuantumNous/new-api/commit/2035a82aeb5414253a728bd937d4b8f97aa99b9b) | pending | 2 | 1 | review required |
| commit-feefe09f2781 | [feefe09f2781](https://github.com/QuantumNous/new-api/commit/feefe09f2781429e039c3f7ba88da7f29492e98c) | pending | 6 | 2 | review required |
| commit-d0cb7347c07b | [d0cb7347c07b](https://github.com/QuantumNous/new-api/commit/d0cb7347c07bc935187c077b960435f3f3b08595) | pending | 2 | 1 | review required |
| commit-8112120673bf | [8112120673bf](https://github.com/QuantumNous/new-api/commit/8112120673bf593eb62e3a297dafca32a4605eb6) | pending | 2 | 0 | review required |
| commit-dfd3cd893044 | [dfd3cd893044](https://github.com/QuantumNous/new-api/commit/dfd3cd8930447780a5733eaaba9a306df595251d) | pending | 1 | 1 | review required |
| commit-5fe8917be728 | [5fe8917be728](https://github.com/QuantumNous/new-api/commit/5fe8917be728a93a29072a9fbd65429ed30d2930) | pending | 2 | 0 | review required |
| commit-56758edf95ec | [56758edf95ec](https://github.com/QuantumNous/new-api/commit/56758edf95ec162033a6b73b553ac789404f87c9) | pending | 22 | 9 | review required |
| commit-1a4166d8e8ba | [1a4166d8e8ba](https://github.com/QuantumNous/new-api/commit/1a4166d8e8ba9802d2ca56fe8ecf0ed5404e80d5) | pending | 13 | 0 | review required |
| commit-b48b74ab77cf | [b48b74ab77cf](https://github.com/QuantumNous/new-api/commit/b48b74ab77cf289ccaaa8d99c067ce50cced4a82) | pending | 6 | 3 | protected files |
| commit-973cf8ef4600 | [973cf8ef4600](https://github.com/QuantumNous/new-api/commit/973cf8ef4600947a4270e95ada7916740fa8264c) | pending | 1 | 0 | review required |
| commit-78bd5b1cbf7b | [78bd5b1cbf7b](https://github.com/QuantumNous/new-api/commit/78bd5b1cbf7bffc462a515a6c6e27567ace7a4d8) | pending | 10 | 0 | review required |
| commit-8ca26d81dfff | [8ca26d81dfff](https://github.com/QuantumNous/new-api/commit/8ca26d81dfff9d0454e0f5d06394c93689fe69b1) | pending | 7 | 2 | review required |
| commit-956019109bb0 | [956019109bb0](https://github.com/QuantumNous/new-api/commit/956019109bb09eb94417fd69ec0a5af0d897ac5a) | pending | 4 | 1 | review required |
| commit-c6bac617d46a | [c6bac617d46a](https://github.com/QuantumNous/new-api/commit/c6bac617d46ad2d63fa7cb692942fe6d21e264f7) | pending | 13 | 7 | review required |
| commit-0908d9e7b911 | [0908d9e7b911](https://github.com/QuantumNous/new-api/commit/0908d9e7b91124e5808f20d2d4b9402c7f2f5fae) | pending | 3 | 2 | review required |
| commit-bb28f2ccad3f | [bb28f2ccad3f](https://github.com/QuantumNous/new-api/commit/bb28f2ccad3f6286a9dc4a0ad9e20f571f57533b) | pending | 35 | 11 | protected files |
| commit-0a532f7f7d1b | [0a532f7f7d1b](https://github.com/QuantumNous/new-api/commit/0a532f7f7d1b507cc6657d6bc140f73d4298e46a) | pending | 3 | 0 | review required |
| commit-0a03088e0e73 | [0a03088e0e73](https://github.com/QuantumNous/new-api/commit/0a03088e0e737e99bc7ca936d3d2276e8d636fb8) | pending | 22 | 1 | review required |
| commit-acb347a197ff | [acb347a197ff](https://github.com/QuantumNous/new-api/commit/acb347a197ff0e22f651be9813b4a1a867342941) | pending | 52 | 14 | review required |
| commit-45094bdf41fb | [45094bdf41fb](https://github.com/QuantumNous/new-api/commit/45094bdf41fb13a5d24bd3e8eb3fdfec6241c59f) | pending | 1 | 0 | review required |
| commit-04c597bf6cce | [04c597bf6cce](https://github.com/QuantumNous/new-api/commit/04c597bf6cce1eea76b9e67a4936e15fb54a5612) | pending | 4 | 0 | review required |
| commit-d10f11ffff20 | [d10f11ffff20](https://github.com/QuantumNous/new-api/commit/d10f11ffff20688b8e3e745268cc0036e896d068) | pending | 11 | 0 | review required |
| commit-0cce21db0283 | [0cce21db0283](https://github.com/QuantumNous/new-api/commit/0cce21db0283304d139bf22588f4dc31bdc5ae67) | pending | 2 | 0 | review required |
| commit-bc823579294f | [bc823579294f](https://github.com/QuantumNous/new-api/commit/bc823579294fbca0064b622c22477d3d04ca1bd8) | pending | 4 | 0 | review required |
| commit-b23471112d8f | [b23471112d8f](https://github.com/QuantumNous/new-api/commit/b23471112d8f7e5848f70f30b798ee417d1ec75c) | pending | 2 | 2 | review required |
| commit-192802ee0691 | [192802ee0691](https://github.com/QuantumNous/new-api/commit/192802ee0691900575a3c7a0cdb90b2d7e703035) | pending | 4 | 3 | protected files |
| commit-6370b2942416 | [6370b2942416](https://github.com/QuantumNous/new-api/commit/6370b29424168039e94d40d610191e7d2e65dbf4) | pending | 11 | 0 | review required |
| commit-7aa3531ef4c2 | [7aa3531ef4c2](https://github.com/QuantumNous/new-api/commit/7aa3531ef4c247ad4891c06cc8ed9d0ffb73fecc) | pending | 2 | 0 | review required |
| commit-e0483122c7d3 | [e0483122c7d3](https://github.com/QuantumNous/new-api/commit/e0483122c7d3693629f0e84473380a8c9506c56a) | pending | 3 | 0 | review required |
| commit-128089fa7ddf | [128089fa7ddf](https://github.com/QuantumNous/new-api/commit/128089fa7ddf41b575206a0feb28940d33793833) | pending | 288 | 77 | protected files |
| commit-1d4328e97417 | [1d4328e97417](https://github.com/QuantumNous/new-api/commit/1d4328e97417a043a161a0dd30a5b129be3ace49) | pending | 1 | 0 | protected files |
| commit-b66674ac23d7 | [b66674ac23d7](https://github.com/QuantumNous/new-api/commit/b66674ac23d79bfd35a90797ca1ac64fc74a2d8f) | pending | 12 | 0 | protected files |
| commit-29509776ef85 | [29509776ef85](https://github.com/QuantumNous/new-api/commit/29509776ef855965fb2cff3978578d891005945e) | pending | 18 | 2 | protected files |
| commit-66c3e142820e | [66c3e142820e](https://github.com/QuantumNous/new-api/commit/66c3e142820e513a328def88f2771f73479275bc) | pending | 3 | 0 | review required |
| commit-3f02244783e0 | [3f02244783e0](https://github.com/QuantumNous/new-api/commit/3f02244783e0bb96b1398601f156625936be2f15) | pending | 4 | 0 | review required |
| commit-ba3b2c015ad8 | [ba3b2c015ad8](https://github.com/QuantumNous/new-api/commit/ba3b2c015ad83e6710c18592b0e3e3ba9df9a67e) | pending | 7 | 0 | review required |
| commit-c8d234baa18b | [c8d234baa18b](https://github.com/QuantumNous/new-api/commit/c8d234baa18b9090b394fc1b8f384eec94f8bfef) | pending | 2 | 0 | review required |
| commit-c6741c36a9a1 | [c6741c36a9a1](https://github.com/QuantumNous/new-api/commit/c6741c36a9a1a729554ad94ee99f21602a214a23) | pending | 29 | 18 | review required |
| commit-1ca79d57e77d | [1ca79d57e77d](https://github.com/QuantumNous/new-api/commit/1ca79d57e77d91c855371b43bb26b151429e18a1) | pending | 10 | 0 | review required |
| commit-0f38a14a5e8e | [0f38a14a5e8e](https://github.com/QuantumNous/new-api/commit/0f38a14a5e8e322ea741478b0e49878be3130e1d) | pending | 7 | 2 | review required |
| commit-036ba376da7d | [036ba376da7d](https://github.com/QuantumNous/new-api/commit/036ba376da7d572a4db3e0be8553aadc28c3616a) | pending | 4 | 0 | review required |
| commit-13c8e90b8f4b | [13c8e90b8f4b](https://github.com/QuantumNous/new-api/commit/13c8e90b8f4b929aa88c2e9616efc462e0068d72) | pending | 2 | 0 | review required |
| commit-5ac75628409d | [5ac75628409d](https://github.com/QuantumNous/new-api/commit/5ac75628409de7e3adde7d36771700714af50853) | pending | 1 | 0 | review required |
| commit-977450a3e810 | [977450a3e810](https://github.com/QuantumNous/new-api/commit/977450a3e810ee9c7d7e98acce1dc4efad90b491) | pending | 4 | 0 | review required |
| commit-b892a6b2e44d | [b892a6b2e44d](https://github.com/QuantumNous/new-api/commit/b892a6b2e44df572698f6cf731ff381f1977a1f2) | pending | 2 | 0 | review required |
| commit-1fb6d07fee01 | [1fb6d07fee01](https://github.com/QuantumNous/new-api/commit/1fb6d07fee0171d5f4a0bd2d2ea2660572b08389) | pending | 8 | 1 | review required |
| commit-06dd76a9d38e | [06dd76a9d38e](https://github.com/QuantumNous/new-api/commit/06dd76a9d38eb0eaf8e900b1264b500250156c4d) | pending | 2 | 0 | review required |
| commit-96275d224bfc | [96275d224bfc](https://github.com/QuantumNous/new-api/commit/96275d224bfcf6b5545ca7e57b3bf90f7c3ad396) | pending | 7 | 1 | review required |
| commit-93ea55f3c310 | [93ea55f3c310](https://github.com/QuantumNous/new-api/commit/93ea55f3c3102f7bc11cd005570c7965f0f22a03) | pending | 4 | 1 | review required |
| commit-4b13d9388ed6 | [4b13d9388ed6](https://github.com/QuantumNous/new-api/commit/4b13d9388ed6f43118707534daad4903e5d22745) | pending | 2 | 0 | review required |
| commit-801456a061fb | [801456a061fb](https://github.com/QuantumNous/new-api/commit/801456a061fb9988957341c4eb5aa8ec20c40de1) | pending | 3 | 0 | review required |
| commit-c10780e054e6 | [c10780e054e6](https://github.com/QuantumNous/new-api/commit/c10780e054e6e518c9d0d8a903526663c9b57d75) | pending | 10 | 1 | review required |
| commit-42e694eb60fe | [42e694eb60fe](https://github.com/QuantumNous/new-api/commit/42e694eb60fe3191795c00d416571ced716eb9c4) | pending | 2 | 0 | review required |
| commit-3827fd83076b | [3827fd83076b](https://github.com/QuantumNous/new-api/commit/3827fd83076b9cbaa3f1d9f9d2957ae653573ed7) | pending | 2 | 0 | review required |
| commit-3f926c3ef497 | [3f926c3ef497](https://github.com/QuantumNous/new-api/commit/3f926c3ef49745be273cd8f753fb691abdd7b836) | pending | 1 | 0 | review required |
| commit-bbf60b49f6a4 | [bbf60b49f6a4](https://github.com/QuantumNous/new-api/commit/bbf60b49f6a4cab709e49e57d45f7911fc7dccae) | pending | 2 | 0 | review required |
| commit-b22a43b6e89e | [b22a43b6e89e](https://github.com/QuantumNous/new-api/commit/b22a43b6e89e5349d8cdbddc88d570b16e15fb1f) | pending | 1 | 1 | review required |
| commit-34b19421237f | [34b19421237f](https://github.com/QuantumNous/new-api/commit/34b19421237fc4695e919cdd799c1b7813023085) | pending | 2 | 1 | review required |
| commit-ba0e478e061e | [ba0e478e061e](https://github.com/QuantumNous/new-api/commit/ba0e478e061ee0c824922d89c05620de878103cf) | pending | 2 | 0 | review required |
| commit-fc08b76f0fb3 | [fc08b76f0fb3](https://github.com/QuantumNous/new-api/commit/fc08b76f0fb3f95e7cb81667e7bf350200147fb3) | pending | 2 | 1 | review required |
| commit-4bfcf2812198 | [4bfcf2812198](https://github.com/QuantumNous/new-api/commit/4bfcf2812198913156a3e88a43e746b61a7b6ef5) | pending | 2 | 0 | review required |
| commit-9948b1ef14ea | [9948b1ef14ea](https://github.com/QuantumNous/new-api/commit/9948b1ef14ea17fb23539386c0cb09018671c955) | pending | 2 | 0 | review required |
| commit-2ed058795bf0 | [2ed058795bf0](https://github.com/QuantumNous/new-api/commit/2ed058795bf0ce968134aa8db18b0dd86b54c30e) | pending | 8 | 1 | protected files |
| commit-2b85a8fc67bd | [2b85a8fc67bd](https://github.com/QuantumNous/new-api/commit/2b85a8fc67bdb6d8d41a25f60f734dd3b9917618) | pending | 1 | 1 | review required |
| commit-767a82b87647 | [767a82b87647](https://github.com/QuantumNous/new-api/commit/767a82b8764778b847462ebecdc231eacd1e282c) | pending | 6 | 0 | review required |
| commit-ba29f0645394 | [ba29f0645394](https://github.com/QuantumNous/new-api/commit/ba29f06453942d9b5c0b32330340b8a000692c4f) | pending | 1 | 0 | review required |
| commit-b8c2dc12660d | [b8c2dc12660d](https://github.com/QuantumNous/new-api/commit/b8c2dc12660dceb2a905008a672f8bc0260c98c8) | pending | 2 | 0 | review required |
| commit-c730add954e1 | [c730add954e1](https://github.com/QuantumNous/new-api/commit/c730add954e1b565656a9320444cac5fee12cfe4) | pending | 2 | 0 | review required |
| commit-82306b3c3594 | [82306b3c3594](https://github.com/QuantumNous/new-api/commit/82306b3c35945bc0ed2c567688afe28d8a281300) | pending | 2 | 0 | review required |
| commit-bf4e1ebfb885 | [bf4e1ebfb885](https://github.com/QuantumNous/new-api/commit/bf4e1ebfb88579930d75d852d2d2bca74cbd5f62) | pending | 2 | 0 | review required |
| commit-7e25797d1468 | [7e25797d1468](https://github.com/QuantumNous/new-api/commit/7e25797d14681e4e16c2b46831450e7d6a5ccc74) | pending | 2 | 0 | review required |
| commit-7f51a97226c4 | [7f51a97226c4](https://github.com/QuantumNous/new-api/commit/7f51a97226c43bc082a065bdc8b40b719c4bdcce) | pending | 1 | 0 | review required |
| commit-1c58f3d161fa | [1c58f3d161fa](https://github.com/QuantumNous/new-api/commit/1c58f3d161fa32755b372d9449f2d5e077d50a5d) | pending | 1 | 0 | review required |
| commit-3fe394cdba28 | [3fe394cdba28](https://github.com/QuantumNous/new-api/commit/3fe394cdba280c00bc566de493b5fc8399b0a7d1) | pending | 2 | 2 | review required |
| commit-9f048107edbb | [9f048107edbb](https://github.com/QuantumNous/new-api/commit/9f048107edbba2156691718126de7c4577614533) | pending | 2 | 0 | review required |
| commit-3c9fb2a0de16 | [3c9fb2a0de16](https://github.com/QuantumNous/new-api/commit/3c9fb2a0de168d88ce5bffbf44acecdc1cdcdd39) | pending | 1 | 0 | review required |
| commit-93cf4ba6fb92 | [93cf4ba6fb92](https://github.com/QuantumNous/new-api/commit/93cf4ba6fb92e3f755fdeb6af9a9484bfa748816) | pending | 0 | 0 | review required |
| commit-dd59854050f2 | [dd59854050f2](https://github.com/QuantumNous/new-api/commit/dd59854050f26f4f67bb3dac5274af5e17013462) | pending | 1 | 0 | review required |
| commit-633b4c5b190d | [633b4c5b190d](https://github.com/QuantumNous/new-api/commit/633b4c5b190dc1af99a2814391354d86948e7943) | pending | 36 | 20 | review required |

## commit-cae3676ec6f4

- Title: feat: glm chanel /v1/responses (#7050)
- Feature/group: feat: glm chanel /v1/responses (#7050)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/zhipu_4v/adaptor.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ba2e9287bb7a

- Title: feat(ollama): passthrough Claude Messages and OpenAI Responses (#7051)
- Feature/group: feat(ollama): passthrough Claude Messages and OpenAI Responses (#7051)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/adaptor.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e468b73915e5

- Title: docs: update PR template and remove PR Check workflow (#7053)
- Feature/group: docs: update PR template and remove PR Check workflow (#7053)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.agents/github/ISSUE.md`, `.agents/github/PR.md`, `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE/en.md`, `.github/workflows/pr-check.yml`, `AGENTS.md`
- Overlap with downstream changes: .github/PULL_REQUEST_TEMPLATE.md, .github/workflows/pr-check.yml, AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-692e8d6ee6a9

- Title: fix(web): restore admin unbinding for built-in providers (#6987)
- Feature/group: fix(web): restore admin unbinding for built-in providers (#6987)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ac381acf4bf4

- Title: fix(billing): 修复时间规则恒真表达式导致倍率全天生效 (#6934)
- Feature/group: fix(billing): 修复时间规则恒真表达式导致倍率全天生效 (#6934)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/lib/__tests__/time-rule-expr.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7037ac15bd8a

- Title: fix(docker): add relaykit go.mod to dev build context (#7072)
- Feature/group: fix(docker): add relaykit go.mod to dev build context (#7072)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `Dockerfile.dev`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-eb48396d5fe9

- Title: feat(task): replace built-in task adaptors with a sandboxed JS plugin system (#7076)
- Feature/group: feat(task): replace built-in task adaptors with a sandboxed JS plugin system (#7076)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `CLAUDE.md`, `THIRD-PARTY-LICENSES.md`, `common/api_type.go`, `common/api_type_task_plugin_test.go`, `common/init.go`, `common/trusted_proxies.go`, `constant/channel.go`, `constant/channel_test.go`, `constant/context_key.go`, `constant/env.go`, `constant/task.go`, `constant/task_test.go`, `controller/billing_option_test.go`, `controller/channel-billing.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_pin_retry_test.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/channel_upstream_update.go`, `controller/log.go`, `controller/model.go`, `controller/option.go`, `controller/plugin_endpoint_test.go`, `controller/plugin_native_e2e_test.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_limiter.go`, `controller/plugin_protocol_limiter_test.go`, `controller/plugin_protocol_sdk_test.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task.go`, `controller/task_generic_test.go`, `controller/task_log_view_test.go`, `controller/task_plugin.go`, `controller/task_plugin_debug.go`, `controller/task_plugin_debug_test.go`, `controller/task_plugin_test.go`, `controller/video_proxy.go`, `controller/video_proxy_gemini.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/channel_constraints.go`, `dto/channel_constraints_test.go`, `dto/plugin_protocol.go`, `dto/task.go`, `dto/task_plugin.go`, `e2e/doc_parse_test.go`, `go.mod`, `go.sum`, `logger/logger.go`, `main.go`, `middleware/auth.go`, `middleware/body_cleanup.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/jimeng_adapter.go`, `middleware/kling_adapter.go`, `middleware/logger.go`, `middleware/task_artifact_access.go`, `middleware/task_artifact_access_test.go`, `middleware/task_plugin.go`, `middleware/task_plugin_origin_task_test.go`, `middleware/task_plugin_test.go`, `middleware/trusted_proxies.go`, `middleware/utils.go`, `model/ability.go`, `model/channel.go`, `model/channel_cache.go`, `model/channel_constraint.go`, `model/channel_constraint_test.go`, `model/log.go`, `model/log_format_test.go`, `model/main.go`, `model/option.go`, `model/option_task_plugin_test.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task.go`, `model/task_cas_test.go`, `model/task_openai_video_test.go`, `model/task_plugin.go`, `model/task_plugin_channel_select_test.go`, `model/task_plugin_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/compile_usage_test.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/task_usage_test.go`, `pkg/billingexpr/types.go`, `pkg/jsplugin/cli.go`, `pkg/jsplugin/cli_test.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/fixture.go`, `pkg/jsplugin/fixture_test.go`, `pkg/jsplugin/protocol_supports_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_disabled_factory_test.go`, `pkg/jsplugin/registry_master_enabled_test.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/request.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/.oxfmtrc.json`, `plugins/.oxlintrc.json`, `plugins/alibaba_responses_test.go`, `plugins/builtin_plugins_test.go`, `plugins/doubao_responses_test.go`, `plugins/embed.go`, `plugins/google_responses_test.go`, `plugins/hailuo_responses_test.go`, `plugins/jimeng_responses_test.go`, `plugins/kling_responses_test.go`, `plugins/sora_responses_test.go`, `plugins/sunoapi_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `plugins/vertex_ai_responses_test.go`, `plugins/video_responses_test_helpers_test.go`, `plugins/vidu_responses_test.go`, `relay/channel/adapter.go`, `relay/channel/api_request.go`, `relay/channel/api_request_test.go`, `relay/channel/minimax/relay-minimax.go`, `relay/channel/replicate/adaptor.go`, `relay/channel/task/ali/adaptor.go`, `relay/channel/task/ali/adaptor_test.go`, `relay/channel/task/ali/constants.go`, `relay/channel/task/doubao/adaptor.go`, `relay/channel/task/doubao/constants.go`, `relay/channel/task/gemini/adaptor.go`, `relay/channel/task/gemini/billing.go`, `relay/channel/task/gemini/dto.go`, `relay/channel/task/gemini/image.go`, `relay/channel/task/hailuo/adaptor.go`, `relay/channel/task/hailuo/constants.go`, `relay/channel/task/hailuo/models.go`, `relay/channel/task/jimeng/adaptor.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth.go`, `relay/channel/task/jsplugin/auth_test.go`, `relay/channel/task/kling/adaptor.go`, `relay/channel/task/sora/adaptor.go`, `relay/channel/task/sora/adaptor_test.go`, `relay/channel/task/sora/constants.go`, `relay/channel/task/suno/adaptor.go`, `relay/channel/task/suno/models.go`, `relay/channel/task/vertex/adaptor.go`, `relay/channel/task/vidu/adaptor.go`, `relay/channel/volcengine/adaptor.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/common/relay_info.go`, `relay/common/relay_utils.go`, `relay/common/relay_utils_test.go`, `relay/constant/relay_mode.go`, `relay/plugin_protocol.go`, `relay/plugin_protocol_test.go`, `relay/relay_adaptor.go`, `relay/relay_adaptor_jsplugin_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `relay/task_platform_error_test.go`, `relay/task_platform_test.go`, `relaykit/dto/channel_settings.go`, `router/api-router.go`, `router/main.go`, `router/plugin-router.go`, `router/plugin_router_test.go`, `router/relay-router.go`, `router/retired_frontend_routes_test.go`, `router/task-plugin-protocol-router.go`, `router/task-router.go`, `router/task_plugin_options_router_test.go`, `router/task_plugin_protocol_router_test.go`, `router/task_router_test.go`, `router/video-router.go`, `router/video_router_test.go`, `router/web-router.go`, `service/authz/authz_test.go`, `service/authz/resources_task_plugin.go`, `service/channel_select.go`, `service/channel_select_test.go`, `service/codex_channel_models.go`, `service/task_artifact_access.go`, `service/task_artifact_access_test.go`, `service/task_artifact_store.go`, `service/task_artifact_store_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/task_plugin_view.go`, `service/task_plugin_view_test.go`, `service/task_polling.go`, `service/task_polling_test.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `setting/system_setting/system_setting_old.go`, `setting/system_setting/task_artifact.go`, `setting/system_setting/task_artifact_store.go`, `setting/system_setting/task_artifact_store_test.go`, `setting/system_setting/task_artifact_test.go`, `setting/task_plugin.go`, `setting/task_plugin_test.go`, `setting/task_pricing_setting/config.go`, `setting/task_pricing_setting/config_test.go`, `types/task_artifact.go`, `web/.gitignore`, `web/bun.lock`, `web/package.json`, `web/rsbuild.config.ts`, `web/scripts/sync-i18n.mjs`, `web/src/components/ai-elements/code-block.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/channel-type-options.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/pricing/__tests__/breakdown-tier-match.test.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-expr.test.ts`, `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-billing-mode-badge.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/pricing/constants.ts`, `web/src/features/pricing/hooks/use-pricing-data.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-mode.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/filters.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/__tests__/task-public-address.test.ts`, `web/src/features/system-settings/general/system-info-section.tsx`, `web/src/features/system-settings/general/task-public-address.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-ratio-table-columns.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/site/index.tsx`, `web/src/features/system-settings/site/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/__tests__/enabled-option.test.ts`, `web/src/features/task-plugins/__tests__/marketplace-panel.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/__tests__/plugin-card.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon.test.ts`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/usage-schema-table.test.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/javascript-viewer.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/marketplace-sources-dialog.tsx`, `web/src/features/task-plugins/components/plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-metadata-card.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/task-plugins/components/plugin-source-picker.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/components/source-diff.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/index.tsx`, `web/src/features/task-plugins/lib/host-protocols.ts`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/features/task-plugins/lib/plugin-icon.ts`, `web/src/features/task-plugins/lib/plugin-url.ts`, `web/src/features/task-plugins/types.ts`, `web/src/features/usage-logs/__tests__/access.test.ts`, `web/src/features/usage-logs/__tests__/artifacts.test.ts`, `web/src/features/usage-logs/__tests__/mobile-layout.test.ts`, `web/src/features/usage-logs/__tests__/task-details.test.ts`, `web/src/features/usage-logs/api.ts`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/task-details-dialog.tsx`, `web/src/features/usage-logs/components/plugin-author-link.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/components/usage-logs-mobile-card.tsx`, `web/src/features/usage-logs/components/usage-logs-provider.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/constants.ts`, `web/src/features/usage-logs/lib/columns.ts`, `web/src/features/usage-logs/lib/query-params.ts`, `web/src/features/usage-logs/lib/task-artifacts.ts`, `web/src/features/usage-logs/lib/task-details.ts`, `web/src/features/usage-logs/lib/task-mobile-layout.ts`, `web/src/features/usage-logs/lib/utils.ts`, `web/src/features/usage-logs/types.ts`, `web/src/hooks/use-sidebar-data.ts`, `web/src/i18n/locales/_reports/_sync-report.json`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/__tests__/localized-text.test.ts`, `web/src/lib/admin-permissions.ts`, `web/src/lib/localized-text.ts`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/task-plugins/index.tsx`
- Overlap with downstream changes: THIRD-PARTY-LICENSES.md, common/api_type.go, constant/channel.go, constant/context_key.go, constant/task.go, controller/channel-test.go, controller/channel.go, controller/log.go, controller/model.go, controller/option.go, controller/relay.go, controller/task.go, controller/video_proxy.go, dto/task.go, go.mod, go.sum, logger/logger.go, main.go, middleware/distributor.go, model/ability.go, model/channel.go, model/channel_cache.go, model/log.go, model/log_format_test.go, model/main.go, model/option.go, model/task.go, model/task_cas_test.go, relay/channel/task/ali/adaptor.go, relay/channel/task/ali/adaptor_test.go, relay/channel/task/doubao/adaptor.go, relay/channel/task/doubao/constants.go, relay/common/relay_info.go, relay/constant/relay_mode.go, relay/relay_adaptor.go, relay/relay_task.go, relaykit/dto/channel_settings.go, router/api-router.go, router/relay-router.go, router/video-router.go, router/web-router.go, service/authz/authz_test.go, service/task_billing.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go, web/rsbuild.config.ts, web/scripts/sync-i18n.mjs, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/system-settings/general/system-info-section.tsx, web/src/features/system-settings/site/index.tsx, web/src/features/system-settings/site/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/api.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/components/usage-logs-mobile-card.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/features/usage-logs/constants.ts, web/src/features/usage-logs/lib/query-params.ts, web/src/features/usage-logs/lib/utils.ts, web/src/features/usage-logs/types.ts, web/src/hooks/use-sidebar-data.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0f2a2075ab07

- Title: fix(relay): 请求参数校验错误返回 HTTP 400 (#6774)
- Feature/group: fix(relay): 请求参数校验错误返回 HTTP 400 (#6774)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/relay.go`
- Overlap with downstream changes: controller/relay.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-98d50d5383a3

- Title: fix(web): recheck setup status after page reload (#6968)
- Feature/group: fix(web): recheck setup status after page reload (#6968)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/routes/__root.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b80d633cf586

- Title: feat(auth): encrypt password login transport
- Feature/group: feat(auth): encrypt password login transport
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/password_crypto.go`, `controller/user.go`, `main.go`, `model/main.go`, `model/password_crypto.go`, `router/api-router.go`, `web/bun.lock`, `web/package.json`, `web/src/features/auth/api.ts`, `web/src/features/auth/lib/password-encryption.ts`
- Overlap with downstream changes: controller/user.go, main.go, model/main.go, router/api-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8454082f930f

- Title: feat(chat): add AQBot preset (#7079)
- Feature/group: feat(chat): add AQBot preset (#7079)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `setting/chat.go`, `web/src/features/chat/lib/chat-links.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-918427d8ab41

- Title: feat(auth): make password encryption opt-in #6743
- Feature/group: feat(auth): make password encryption opt-in #6743
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/constants.go`, `common/init.go`, `controller/misc.go`, `controller/user.go`, `main.go`, `web/src/features/auth/api.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/types.ts`
- Overlap with downstream changes: .env.example, common/constants.go, controller/misc.go, controller/user.go, main.go, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/auth/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6c22550ea325

- Title: feat(task): resolve channel-mapped aliases and case variants for plugin models
- Feature/group: feat(task): resolve channel-mapped aliases and case variants for plugin models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/billing_option_test.go`, `controller/option.go`, `controller/plugin_protocol_test.go`, `docs/plugin-api/v1.md`, `middleware/task_plugin.go`, `middleware/task_plugin_model.go`, `middleware/task_plugin_test.go`, `model/channel_cache.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_model_alias.go`, `pkg/jsplugin/model_fold.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vidu/plugin.js`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`
- Overlap with downstream changes: controller/option.go, model/channel_cache.go, relay/relay_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-66031a09d99f

- Title: fix(model): disable PostgreSQL prepared statements for pooler compatibility
- Feature/group: fix(model): disable PostgreSQL prepared statements for pooler compatibility
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`, `model/gorm_logger.go`, `model/gorm_logger_test.go`, `model/main.go`
- Overlap with downstream changes: go.mod, go.sum, model/gorm_logger.go, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0bee5d441029

- Title: fix(ali): honor image response format (#5513) (#7048)
- Feature/group: fix(ali): honor image response format (#5513) (#7048)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-dc4732cfed71

- Title: feat(web): factory task plugins update only with the system
- Feature/group: feat(web): factory task plugins update only with the system
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b5b94bc685fd

- Title: fix(subscription): 无有效订阅时前端如实显示「仅用订阅」偏好 (#6222) (#7086)
- Feature/group: fix(subscription): 无有效订阅时前端如实显示「仅用订阅」偏好 (#6222) (#7086)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/wallet/components/subscription-plans-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1751f43ee07e

- Title: fix(sqlite): enable WAL + working busy timeout + _txlock=immediate to stop concurrent write lockouts (#7030)
- Feature/group: fix(sqlite): enable WAL + working busy timeout + _txlock=immediate to stop concurrent write lockouts (#7030)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/database.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6eb6f35ed211

- Title: fix(model): return string from JSON column Valuers for pg simple protocol
- Feature/group: fix(model): return string from JSON column Valuers for pg simple protocol
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/channel.go`, `model/json_column_test.go`, `model/main.go`, `model/prefill_group.go`, `model/task.go`
- Overlap with downstream changes: model/channel.go, model/main.go, model/task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b518d0033b67

- Title: fix(relay): bound the wait for upstream response headers (fixes unbounded heap growth → OOM) (#6949)
- Feature/group: fix(relay): bound the wait for upstream response headers (fixes unbounded heap growth → OOM) (#6949)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.env.example`, `README.md`, `common/constants.go`, `common/init.go`, `service/http_client.go`
- Overlap with downstream changes: .env.example, README.md, common/constants.go, service/http_client.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-74158715cde6

- Title: fix initialize database
- Feature/group: fix initialize database
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-69a41eeadc81

- Title: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate (#7100)
- Feature/group: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate (#7100)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/prefill_group_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2bf0820f4b89

- Title: Revert "fix(model): drop leftover prefill_groups unique constraints before Au…" (#7101)
- Feature/group: Revert "fix(model): drop leftover prefill_groups unique constraints before Au…" (#7101)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/prefill_group_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2b6f1dfefbe2

- Title: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate
- Feature/group: fix(model): drop leftover prefill_groups unique constraints before AutoMigrate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `model/main.go`, `model/prefill_group_migration.go`, `model/prefill_group_migration_test.go`
- Overlap with downstream changes: AGENTS.md, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8c8c4153d4b8

- Title: fix(log): preserve quota in usage statistics (#7108)
- Feature/group: fix(log): preserve quota in usage statistics (#7108)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/log.go`
- Overlap with downstream changes: model/log.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-27ff6a8767e7

- Title: fix(model): migrate legacy token key constraints
- Feature/group: fix(model): migrate legacy token key constraints
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/token_migration.go`, `model/token_migration_test.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-67a0585d0f25

- Title: fix(docs): correct Video API links across localized READMEs (#7116)
- Feature/group: fix(docs): correct Video API links across localized READMEs (#7116)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b7017c251bad

- Title: fix(model): do not treat no-op system task state writes as lock loss (#7135)
- Feature/group: fix(model): do not treat no-op system task state writes as lock loss (#7135)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/system_task.go`, `model/system_task_test.go`
- Overlap with downstream changes: model/system_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0ed497f066a6

- Title: feat(relay): hosted-tool conversion fidelity, reasoning normalization, and billing usage integrity (#7137)
- Feature/group: feat(relay): hosted-tool conversion fidelity, reasoning normalization, and billing usage integrity (#7137)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.gitignore`, `controller/channel-test.go`, `controller/relay.go`, `controller/relay_count_tokens_test.go`, `model/channel.go`, `relay/channel/aws/adaptor.go`, `relay/channel/aws/relay_aws_test.go`, `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/adaptor.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/newapi/adaptor.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/chat_via_responses_test.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/channel/sub2api/adaptor_test.go`, `relay/channel/vertex/adaptor.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/chat_completions_via_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/claude_handler.go`, `relay/common/conversion_diagnostics.go`, `relay/common/override.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relay/common/tool_usage.go`, `relay/compatible_handler.go`, `relay/convert_request_error.go`, `relay/convert_request_error_test.go`, `relay/gemini_handler.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/helper/reasoning_suffix.go`, `relay/helper/reasoning_suffix_test.go`, `relay/responses_handler.go`, `relaykit/README.md`, `relaykit/dto/billing_usage.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/reasoning_state.go`, `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/reasonmap/reasonmap.go`, `relaykit/relayconvert/claude_default_max_tokens_test.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/convmeta/meta_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/golden_test.go`, `relaykit/relayconvert/internal/claude_messages/citations.go`, `relaykit/relayconvert/internal/claude_messages/stream_billing_usage_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_hosted_stream.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_req.go`, `relaykit/relayconvert/internal/gemini_chat/grounding.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_responses_hosted_stream.go`, `relaykit/relayconvert/internal/oai_chat/citations.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_stream_resp.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/claude/usage.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/hosted_values.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/internal/toolconv/response.go`, `relaykit/relayconvert/internal/toolconv/response_artifacts.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_compat.go`, `relaykit/relayconvert/response_registry.go`, `relaykit/relayconvert/response_registry_test.go`, `relaykit/relayconvert/terminal_stream_test.go`, `relaykit/relayconvert/testdata/golden/request/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/request/claude_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/request/claude_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/request/gemini_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/request/openai_responses_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/request/openai_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/gemini_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_responses_to_openai.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_claude.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/response/openai_to_openai_responses.golden.json`, `relaykit/relayconvert/testdata/golden/stream/claude_to_gemini.golden.json`, `relaykit/relayconvert/testdata/golden/stream/gemini_to_claude.golden.json`, `relaykit/relayconvert/text_converter_registry.go`, `relaykit/relayconvert/text_converter_registry_test.go`, `relaykit/relayconvert/tool_loss_policy_test.go`, `relaykit/types/conversion.go`, `router/relay-router.go`, `router/relay_router_test.go`, `service/billing_session.go`, `service/billing_usage.go`, `service/log_info_generate.go`, `service/quota.go`, `service/request_converter.go`, `service/response_converter.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/token_counter.go`, `setting/model_setting/global.go`, `setting/ratio_setting/model_ratio.go`, `setting/reasoning/suffix.go`
- Overlap with downstream changes: .gitignore, controller/channel-test.go, controller/relay.go, model/channel.go, relay/channel/openai/adaptor.go, relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/chat_completions_via_responses.go, relay/chat_completions_via_responses_test.go, relay/claude_handler.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/responses_handler.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, relaykit/dto/openai_request.go, router/relay-router.go, router/relay_router_test.go, service/billing_session.go, service/quota.go, service/text_quota.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bbd97446c260

- Title: fix(relay): follow-up billing integrity and conversion completions (#7170)
- Feature/group: fix(relay): follow-up billing integrity and conversion completions (#7170)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/openai/helper.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/billing_usage.go`, `relaykit/dto/billing_usage_test.go`, `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`
- Overlap with downstream changes: relay/channel/openai/helper.go, relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-aece11d2f7f0

- Title: feat(plugin): add MiniMax-H3 /v2 video generation to the hailuo task … (#7168)
- Feature/group: feat(plugin): add MiniMax-H3 /v2 video generation to the hailuo task … (#7168)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/relay_task.go`, `service/task_polling.go`
- Overlap with downstream changes: relay/relay_task.go, service/task_polling.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d8ca0ed0bb59

- Title: chore: let owners use human PR templates
- Feature/group: chore: let owners use human PR templates
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-73afad588ca7

- Title: fix(plugin): account for MiniMax-H3 input media usage (#7171)
- Feature/group: fix(plugin): account for MiniMax-H3 input media usage (#7171)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-057f71c2336c

- Title: fix(logs): isolate privileged metadata
- Feature/group: fix(logs): isolate privileged metadata
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.gitignore`, `controller/channel-test.go`, `controller/channel_test_internal_test.go`, `controller/log.go`, `controller/relay.go`, `controller/relay_error_log_test.go`, `model/log.go`, `model/log_format_test.go`, `model/log_other.go`, `model/log_other_test.go`, `relay/convert_request_error_test.go`, `service/billing_usage.go`, `service/channel_affinity.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/quota_saturation_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/violation_fee.go`, `web/src/features/usage-logs/components/__tests__/reject-reason.test.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/types.ts`
- Overlap with downstream changes: .gitignore, controller/channel-test.go, controller/log.go, controller/relay.go, model/log.go, model/log_format_test.go, service/midjourney.go, service/task_billing.go, service/task_billing_test.go, service/text_quota.go, service/violation_fee.go, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-219c9e06341f

- Title: 优化匿名冷启动与公开内容接口的重复回源请求 (#7166)
- Feature/group: 优化匿名冷启动与公开内容接口的重复回源请求 (#7166)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/misc.go`, `controller/revalidated_response.go`, `docs/authentication.md`, `service/auth_session.go`, `web/src/features/about/api.ts`, `web/src/features/home/api.ts`, `web/src/features/legal/api.ts`, `web/src/lib/api.ts`, `web/src/lib/auth-session.ts`, `web/src/lib/session-hint.ts`, `web/src/routes/(auth)/sign-in.tsx`, `web/src/routes/_authenticated/route.tsx`
- Overlap with downstream changes: controller/misc.go, service/auth_session.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9f506dd7f905

- Title: refactor(logs): simplify LogOther projection and dedupe sensitive keys
- Feature/group: refactor(logs): simplify LogOther projection and dedupe sensitive keys
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/log_format_test.go`, `model/log_other.go`, `model/log_other_test.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/text_quota_test.go`
- Overlap with downstream changes: model/log_format_test.go, service/task_billing.go, service/task_billing_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9df450fe54e1

- Title: feat(task): give polling hooks a real query context, host HTTP classification, and bounded poll failures
- Feature/group: feat(task): give polling hooks a real query context, host HTTP classification, and bounded poll failures
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/init.go`, `constant/env.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `model/log_other.go`, `model/task.go`, `model/task_cas_test.go`, `pkg/jsplugin/cli.go`, `pkg/jsplugin/cli_test.go`, `plugins/hailuo_responses_test.go`, `plugins/jimeng_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `plugins/veo_poll_test.go`, `relay/channel/adapter.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/common/relay_info.go`, `relay/relay_task.go`, `service/task_billing_test.go`, `service/task_plugin_view_test.go`, `service/task_polling.go`, `service/task_polling_test.go`
- Overlap with downstream changes: .env.example, controller/relay.go, model/task.go, model/task_cas_test.go, relay/common/relay_info.go, relay/relay_task.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-36dbbf0f77e7

- Title: fix: keep ETag valid across different JSON packages
- Feature/group: fix: keep ETag valid across different JSON packages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/etag.go`, `controller/misc.go`, `controller/revalidated_response.go`
- Overlap with downstream changes: controller/misc.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8f5ab8e4048a

- Title: fix(ci): resolve release version from trigger tag
- Feature/group: fix(ci): resolve release version from trigger tag
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/workflows/release.yml`
- Overlap with downstream changes: .github/workflows/release.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-32c261923a97

- Title: fix(task): explain 503 when a plugin-claimed model has no channel
- Feature/group: fix(task): explain 503 when a plugin-claimed model has no channel
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`
- Overlap with downstream changes: i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3a9f41ee85cc

- Title: fix: temp disable /messages/count_tokens
- Feature/group: fix: temp disable /messages/count_tokens
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `router/relay-router.go`, `router/relay_router_test.go`
- Overlap with downstream changes: router/relay-router.go, router/relay_router_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7c044d7c5c2d

- Title: feat(relay): explicit @ model modifiers and canonical billing identity
- Feature/group: feat(relay): explicit @ model modifiers and canonical billing identity
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel-test.go`, `controller/model.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `model/channel_cache.go`, `model/channel_satisfy.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/openai/adaptor.go`, `relay/claude_handler.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relay/compatible_handler.go`, `relay/convert_request_error_test.go`, `relay/gemini_handler.go`, `relay/helper/model_mapped.go`, `relay/helper/model_modifier.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/helper/reasoning_suffix.go`, `relay/helper/reasoning_suffix_test.go`, `relay/responses_handler.go`, `relaykit/relayconvert/internal/convdiag/collector.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `setting/model_setting/global.go`, `setting/model_setting/global_test.go`, `setting/ratio_setting/matching_test.go`, `setting/ratio_setting/model_ratio.go`, `setting/reasoning/suffix.go`, `setting/reasoning/suffix_test.go`, `web/src/features/system-settings/models/global-settings-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/channel-test.go, controller/model.go, middleware/distributor.go, model/channel_cache.go, relay/channel/openai/adaptor.go, relay/claude_handler.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/helper/model_mapped.go, relay/responses_handler.go, setting/ratio_setting/model_ratio.go, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6b659fd61c50

- Title: fix(relay): preserve reasoning effort without implicit remapping
- Feature/group: fix(relay): preserve reasoning effort without implicit remapping
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/deepseek/adaptor.go`, `relay/channel/openai/adaptor.go`, `relay/channel/volcengine/adaptor.go`, `relay/channel/xai/adaptor.go`, `relay/common/relay_info.go`, `relay/helper/model_modifier.go`, `relay/helper/reasoning_suffix.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d5803532bdcc

- Title: docs: require expression pricing and consolidated tests
- Feature/group: docs: require expression pricing and consolidated tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-eb99ab1b4034

- Title: feat(billing): add built-in expression pricing for gpt-6-astra
- Feature/group: feat(billing): add built-in expression pricing for gpt-6-astra
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/option.go`, `relay/channel/openai/constant.go`, `setting/billing_setting/builtin_billing.go`, `setting/billing_setting/builtin_billing_test.go`, `setting/billing_setting/tiered_billing.go`
- Overlap with downstream changes: controller/option.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2cf177ac487e

- Title: perf(common): 批量复制 RawMessage，优化请求深拷贝 (#7221)
- Feature/group: perf(common): 批量复制 RawMessage，优化请求深拷贝 (#7221)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/copy.go`, `relay/request_clone_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-49ec46966825

- Title: fix(relay): apply model-specific OpenAI chat capabilities (#7211)
- Feature/group: fix(relay): apply model-specific OpenAI chat capabilities (#7211)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel_test_request_test.go`, `relay/channel/openai/adaptor.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d8cb177440ce

- Title: feat(security): add access token management and audit logs
- Feature/group: feat(security): add access token management and audit logs
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/billing_option_test.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_test_internal_test.go`, `controller/model_list_test.go`, `controller/relay.go`, `controller/secure_verification.go`, `controller/subscription.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `go.mod`, `go.sum`, `middleware/audit.go`, `middleware/auth.go`, `middleware/auth_test.go`, `model/audit_log.go`, `model/audit_other.go`, `model/log.go`, `model/main.go`, `model/user.go`, `router/api-router.go`, `router/main.go`, `router/plugin_router_test.go`, `router/web-router.go`, `service/authz/authz_test.go`, `service/authz/resources_audit.go`, `web/src/components/layout/components/mobile-drawer.tsx`, `web/src/components/profile-dropdown.tsx`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/profile/__tests__/settings.test.tsx`, `web/src/features/profile/api.ts`, `web/src/features/profile/components/__tests__/login-session-utils.test.ts`, `web/src/features/profile/components/dialogs/access-token-dialog.tsx`, `web/src/features/profile/components/dialogs/change-password-dialog.tsx`, `web/src/features/profile/components/dialogs/delete-account-dialog.tsx`, `web/src/features/profile/components/dialogs/email-bind-dialog.tsx`, `web/src/features/profile/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/profile/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/profile/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/profile/components/login-session-dialogs.tsx`, `web/src/features/profile/components/login-session-item.tsx`, `web/src/features/profile/components/login-session-utils.ts`, `web/src/features/profile/components/login-sessions-card.tsx`, `web/src/features/profile/components/passkey-card.tsx`, `web/src/features/profile/components/profile-security-card.tsx`, `web/src/features/profile/components/profile-settings-card.tsx`, `web/src/features/profile/components/sidebar-modules-card.tsx`, `web/src/features/profile/components/tabs/account-bindings-tab.tsx`, `web/src/features/profile/components/tabs/notification-tab.tsx`, `web/src/features/profile/components/two-fa-card.tsx`, `web/src/features/profile/hooks/index.ts`, `web/src/features/profile/hooks/use-access-token.ts`, `web/src/features/profile/hooks/use-two-fa.ts`, `web/src/features/profile/index.tsx`, `web/src/features/profile/lib/user-settings.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/page.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/__tests__/login-session-utils.test.ts`, `web/src/features/security/components/__tests__/privacy-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/components/account-action-card.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/access-token-dialog.tsx`, `web/src/features/security/components/dialogs/change-password-dialog.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/email-bind-dialog.tsx`, `web/src/features/security/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/security/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/security/components/login-session-dialogs.tsx`, `web/src/features/security/components/login-session-item.tsx`, `web/src/features/security/components/login-session-utils.ts`, `web/src/features/security/components/login-sessions-card.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/privacy-card.tsx`, `web/src/features/security/components/two-fa-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-two-fa.ts`, `web/src/features/security/index.tsx`, `web/src/features/system-settings/maintenance/config.ts`, `web/src/features/system-settings/maintenance/sidebar-modules-section.tsx`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/api.ts`, `web/src/features/usage-logs/audit/components/audit-detail-fields.tsx`, `web/src/features/usage-logs/audit/components/audit-detail-value.tsx`, `web/src/features/usage-logs/audit/components/audit-log-columns.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/audit/index.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/components/__tests__/log-type-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/log-detail-layout.tsx`, `web/src/features/usage-logs/constants.ts`, `web/src/features/usage-logs/index.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/hooks/__tests__/sidebar-config.test.tsx`, `web/src/hooks/use-sidebar-config.ts`, `web/src/hooks/use-sidebar-data.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/__tests__/http-cache.test.ts`, `web/src/lib/admin-permissions.ts`, `web/src/lib/auth-session.ts`, `web/src/lib/http-client.ts`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/security/index.tsx`, `web/src/routes/_authenticated/usage-logs/audit.tsx`
- Overlap with downstream changes: controller/audit.go, controller/model_list_test.go, controller/relay.go, controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, go.mod, go.sum, model/log.go, model/main.go, model/user.go, router/api-router.go, router/web-router.go, service/authz/authz_test.go, web/src/components/profile-dropdown.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/profile/components/__tests__/login-session-utils.test.ts, web/src/features/profile/components/login-session-utils.ts, web/src/features/profile/components/tabs/account-bindings-tab.tsx, web/src/features/profile/index.tsx, web/src/features/system-settings/maintenance/config.ts, web/src/features/system-settings/maintenance/sidebar-modules-section.tsx, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/compact-date-time-range-picker.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/constants.ts, web/src/features/usage-logs/index.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/hooks/use-sidebar-config.ts, web/src/hooks/use-sidebar-data.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts, web/src/lib/http-client.ts, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9a8674425c5a

- Title: fix(db): avoid redundant schema migrations on restart
- Feature/group: fix(db): avoid redundant schema migrations on restart
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `go.mod`, `go.sum`, `model/main.go`, `model/migration_dialector.go`, `model/migration_dialector_test.go`, `model/user_session_migration_test.go`
- Overlap with downstream changes: AGENTS.md, go.mod, go.sum, model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-45c3fbe8aeb0

- Title: fix(security): bind verification proofs to sessions and actions
- Feature/group: fix(security): bind verification proofs to sessions and actions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `common/json.go`, `common/json_test.go`, `controller/channel.go`, `controller/oauth.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `middleware/auth_test.go`, `middleware/logger.go`, `middleware/secure_verification.go`, `model/auth_flow.go`, `model/passkey.go`, `model/twofa.go`, `model/twofa_enrollment.go`, `model/user_authentication_test.go`, `oauth/discord.go`, `oauth/generic.go`, `oauth/github.go`, `oauth/linuxdo.go`, `oauth/oidc.go`, `router/api-router.go`, `service/auth_token.go`, `service/auth_token_test.go`, `service/oauth_verification.go`, `service/passkey/session.go`, `service/security_verification.go`, `service/twofa.go`, `web/AGENTS.md`, `web/src/features/auth/api.ts`, `web/src/features/auth/components/oauth-callback-screen.tsx`, `web/src/features/auth/constants.ts`, `web/src/features/auth/lib/__tests__/oauth-callback-mode.test.ts`, `web/src/features/auth/lib/__tests__/oauth-popup.test.ts`, `web/src/features/auth/lib/oauth-callback-mode.ts`, `web/src/features/auth/lib/oauth-popup.ts`, `web/src/features/auth/passkey/api.ts`, `web/src/features/auth/passkey/hooks/use-passkey-management.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/__tests__/channel-key-disclosure.test.tsx`, `web/src/features/channels/hooks/use-channel-key-disclosure.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/two-fa-card.tsx`, `web/src/features/security/hooks/use-two-fa-setup.ts`, `web/src/features/security/hooks/use-two-fa.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/api.ts`, `web/src/lib/http-client.ts`, `web/src/lib/oauth.ts`, `web/src/lib/passkey.ts`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.ts`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: AGENTS.md, controller/channel.go, controller/secure_verification.go, model/auth_flow.go, model/twofa.go, model/user_authentication_test.go, router/api-router.go, service/auth_token.go, web/src/features/auth/constants.ts, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/http-client.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3e84ec0ab823

- Title: feat(auth): migrate Telegram to unified OAuth
- Feature/group: feat(auth): migrate Telegram to unified OAuth
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/misc.go`, `controller/oauth.go`, `controller/option.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/telegram.go`, `controller/telegram_test.go`, `controller/twofa.go`, `go.mod`, `go.sum`, `model/custom_oauth_provider.go`, `model/external_identity_claim.go`, `oauth/registry.go`, `oauth/telegram.go`, `oauth/types.go`, `router/api-router.go`, `service/security_verification.go`, `setting/system_setting/telegram.go`, `web/src/features/auth/api.test.ts`, `web/src/features/auth/api.ts`, `web/src/features/auth/components/oauth-providers.tsx`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/lib/__tests__/oauth-popup.test.ts`, `web/src/features/auth/lib/oauth-callback-mode.ts`, `web/src/features/auth/lib/oauth-popup.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/types.ts`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/system-settings/auth/index.tsx`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/features/system-settings/auth/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/server-error-message.ts`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: controller/channel.go, controller/misc.go, controller/option.go, controller/secure_verification.go, go.mod, go.sum, model/external_identity_claim.go, router/api-router.go, web/src/features/auth/components/oauth-providers.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/types.ts, web/src/features/system-settings/auth/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a8729b5c3709

- Title: feat(security): require verification for access token management
- Feature/group: feat(security): require verification for access token management
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/security_enrollment_test.go`, `controller/user.go`, `router/api-router.go`, `service/security_verification.go`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/lib/secure-verification.ts`
- Overlap with downstream changes: controller/user.go, router/api-router.go, web/src/features/auth/secure-verification/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0973dc2b8f55

- Title: feat(security): harden account binding and password changes
- Feature/group: feat(security): harden account binding and password changes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.env.example`, `common/account_password.go`, `common/crypto.go`, `common/password_crypto.go`, `controller/audit.go`, `controller/auth_flow_test.go`, `controller/custom_oauth.go`, `controller/email_binding.go`, `controller/misc.go`, `controller/model_list_test.go`, `controller/oauth.go`, `controller/passkey.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/setup.go`, `controller/telegram_test.go`, `controller/user.go`, `controller/wechat.go`, `middleware/secure_verification.go`, `model/account_security.go`, `model/auth_flow.go`, `model/email_binding.go`, `model/user.go`, `router/api-router.go`, `service/account_security.go`, `service/auth_session.go`, `service/email_binding.go`, `service/security_verification.go`, `web/src/features/auth/api.ts`, `web/src/features/auth/constants.ts`, `web/src/features/auth/lib/password-encryption.ts`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/auth/types.ts`, `web/src/features/profile/api.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/__tests__/enrollment.test.tsx`, `web/src/features/security/components/account-action-card.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/security/components/dialogs/change-password-dialog.tsx`, `web/src/features/security/components/dialogs/email-bind-dialog.tsx`, `web/src/features/security/components/dialogs/wechat-bind-dialog.tsx`, `web/src/features/security/hooks/use-account-security.ts`, `web/src/features/security/index.tsx`, `web/src/features/setup/setup-wizard.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/password-policy.ts`, `web/src/stores/auth-store.ts`
- Overlap with downstream changes: .env.example, controller/audit.go, controller/misc.go, controller/model_list_test.go, controller/secure_verification.go, controller/user.go, model/auth_flow.go, model/user.go, router/api-router.go, service/auth_session.go, web/src/features/auth/constants.ts, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/stores/auth-store.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3f8a50cf8877

- Title: feat(audit): complete token and quota operation records
- Feature/group: feat(audit): complete token and quota operation records
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `constant/context_key.go`, `controller/audit.go`, `controller/subscription.go`, `controller/token.go`, `controller/token_test.go`, `controller/user.go`, `controller/user_manage_test.go`, `controller/user_quota.go`, `middleware/audit.go`, `model/log.go`, `model/quota_reserve.go`, `model/user_quota_adjustment.go`, `router/api-router.go`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/components/audit-detail-fields.tsx`, `web/src/features/usage-logs/audit/components/audit-log-columns.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/components/__tests__/quota-adjustment.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/lib/quota-audit-operation.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: AGENTS.md, constant/context_key.go, controller/audit.go, controller/user.go, controller/user_manage_test.go, model/log.go, router/api-router.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-521cebf585ef

- Title: fix(dashboard): simplify completed setup guide
- Feature/group: fix(dashboard): simplify completed setup guide
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/components/overview/__tests__/setup-guide.test.tsx`, `web/src/features/dashboard/components/overview/overview-dashboard.tsx`, `web/src/features/dashboard/index.tsx`
- Overlap with downstream changes: web/src/features/dashboard/components/overview/overview-dashboard.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6f2333990613

- Title: feat(auth): unify login verification and secure account deletion
- Feature/group: feat(auth): unify login verification and secure account deletion
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/auth_flow_test.go`, `controller/auth_session_test.go`, `controller/login_verification.go`, `controller/model_list_test.go`, `controller/passkey.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `model/account_security.go`, `model/auth_flow.go`, `model/login_verification.go`, `model/twofa.go`, `model/user.go`, `model/user_session.go`, `router/api-router.go`, `service/auth_session.go`, `service/login_verification.go`, `service/passkey/session.go`, `service/security_verification.go`, `web/src/features/auth/hooks/use-auth-redirect.ts`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/otp/__tests__/login-verification.test.tsx`, `web/src/features/auth/otp/components/otp-form.tsx`, `web/src/features/auth/otp/index.tsx`, `web/src/features/auth/secure-verification/__tests__/verification-api.test.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/auth/types.ts`, `web/src/features/profile/api.ts`, `web/src/features/profile/types.ts`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-backup-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-disable-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/api.ts`, `web/src/lib/auth-session.test.ts`, `web/src/lib/passkey.ts`, `web/src/routes/oauth/$provider.tsx`, `web/src/stores/auth-store.ts`
- Overlap with downstream changes: controller/audit.go, controller/model_list_test.go, controller/user.go, model/auth_flow.go, model/twofa.go, model/user.go, router/api-router.go, service/auth_session.go, web/src/features/auth/otp/components/otp-form.tsx, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/auth/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/routes/oauth/$provider.tsx, web/src/stores/auth-store.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0c76e4dae77a

- Title: feat(models): rework model/vendor management and pricing
- Feature/group: feat(models): rework model/vendor management and pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_meta.go`, `controller/model_pricing_config.go`, `controller/model_sync.go`, `controller/ratio_config.go`, `controller/ratio_sync.go`, `controller/ratio_sync_test.go`, `controller/vendor_meta.go`, `model/ability.go`, `model/channel_cache.go`, `model/model_meta.go`, `model/model_metadata_sync.go`, `model/model_pricing_config.go`, `model/option.go`, `model/pricing_default.go`, `model/vendor_management.go`, `model/vendor_meta.go`, `router/api-router.go`, `setting/billing_setting/tiered_billing.go`, `setting/ratio_setting/model_ratio.go`, `web/src/components/data-table/core/badge-list-cell.tsx`, `web/src/components/data-table/layout/data-table-page.tsx`, `web/src/components/data-table/layout/mobile-card-list.tsx`, `web/src/components/data-table/toolbar/bulk-actions.tsx`, `web/src/components/data-table/toolbar/view-mode-toggle.tsx`, `web/src/components/layout/components/section-page-layout.tsx`, `web/src/components/lobe-icon-field.tsx`, `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/model-pricing/__tests__/pricing.test.ts`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/__tests__/metadata-sync.test.tsx`, `web/src/features/models/__tests__/model-deletion.test.tsx`, `web/src/features/models/__tests__/vendor-management.test.tsx`, `web/src/features/models/api.ts`, `web/src/features/models/components/data-table-bulk-actions.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/dialogs/create-deployment-drawer.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/dialogs/price-sync-dialog.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/features/models/components/dialogs/upstream-conflict-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-operation-dialog.tsx`, `web/src/features/models/components/dialogs/vendors-management-dialog.tsx`, `web/src/features/models/components/dialogs/view-logs-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/model-connections.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-dialogs.tsx`, `web/src/features/models/components/models-primary-buttons.tsx`, `web/src/features/models/components/models-provider.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/components/vendor-linked-models.tsx`, `web/src/features/models/components/vendors-table.tsx`, `web/src/features/models/constants.ts`, `web/src/features/models/index.tsx`, `web/src/features/models/lib/model-actions.ts`, `web/src/features/models/section-registry.tsx`, `web/src/features/models/types.ts`, `web/src/features/models/vendor-api.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/pricing-controls.test.tsx`, `web/src/features/pricing/components/loading-skeleton.tsx`, `web/src/features/pricing/components/model-billing-mode-badge.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/pricing/components/pricing-toolbar.tsx`, `web/src/features/pricing/index.tsx`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/profile/components/language-preferences-card.tsx`, `web/src/features/subscriptions/components/dialogs/subscription-purchase-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/system-settings/__tests__/pricing-sync.test.tsx`, `web/src/features/system-settings/auth/custom-oauth/components/preset-selector.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/system-settings/models/channel-selector-dialog.tsx`, `web/src/features/system-settings/models/conflict-confirm-dialog.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/group-special-usable-editor.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/upstream-price-cells.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync-helpers.ts`, `web/src/features/system-settings/models/upstream-ratio-sync-table.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/lobe-icon.tsx`, `web/src/routes/_authenticated/models/$section.tsx`, `web/vitest.config.ts`
- Overlap with downstream changes: model/ability.go, model/channel_cache.go, model/option.go, router/api-router.go, setting/ratio_setting/model_ratio.go, web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/pricing/index.tsx, web/src/features/system-settings/types.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3b4652269a6a

- Title: feat(ali): support wan3.0 all-in-one video models
- Feature/group: feat(ali): support wan3.0 all-in-one video models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_wan3_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6e10f9bc927a

- Title: fix(relay): preserve Kimi K3 dynamic tool loading messages
- Feature/group: fix(relay): preserve Kimi K3 dynamic tool loading messages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/moonshot/constants.go`, `relay/chat_completions_via_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/compatible_handler.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`, `relaykit/relayconvert/kitutil/json_test.go`
- Overlap with downstream changes: relay/channel/moonshot/constants.go, relay/chat_completions_via_responses.go, relay/chat_completions_via_responses_test.go, relay/compatible_handler.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: True
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7bbe85bcb095

- Title: refactor(json): route JSON helpers through a host-injectable codec
- Feature/group: refactor(json): route JSON helpers through a host-injectable codec
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/json.go`, `common/json_test.go`, `relaykit/dto/values.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/kitutil/json.go`
- Overlap with downstream changes: relaykit/dto/values.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-387a40914853

- Title: fix(web): keep drawer popups interactive and shim storage in tests
- Feature/group: fix(web): keep drawer popups interactive and shim storage in tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/__tests__/portal-container.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/components/ui/drawer.tsx`, `web/src/components/ui/portal-container.ts`, `web/src/components/ui/select.tsx`, `web/src/test-setup.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-5c7cca015525

- Title: fix(perf): return hourly success-rate series for model health bar
- Feature/group: fix(perf): return hourly success-rate series for model health bar
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/types.go`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6298b0f32384

- Title: fix(plugin): suppress factory layer when disabling an overridden task plugin
- Feature/group: fix(plugin): suppress factory layer when disabling an overridden task plugin
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-92bc7ff73c5e

- Title: fix(plugins): make sunoapi alias-safe and lock alias echo across built-ins
- Feature/group: fix(plugins): make sunoapi alias-safe and lock alias echo across built-ins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/builtin_plugins_test.go`, `plugins/sunoapi_responses_test.go`, `plugins/tasks/sunoapi/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-210734bb73bc

- Title: refactor(task): remove the custom-plugin layer switch
- Feature/group: refactor(task): remove the custom-plugin layer switch
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/init.go`, `constant/task.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/option.go`, `model/option_task_plugin_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_disabled_factory_test.go`, `pkg/jsplugin/registry_master_enabled_test.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`
- Overlap with downstream changes: constant/task.go, model/option.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bee45b58a3c0

- Title: fix(web): switch the pricing card grid to three columns at xl
- Feature/group: fix(web): switch the pricing card grid to three columns at xl
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/loading-skeleton.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-75e533209490

- Title: feat(pricing): support site currency in pricing editors
- Feature/group: feat(pricing): support site currency in pricing editors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/currency.ts`, `web/src/features/model-pricing/pricing-amount-input.tsx`, `web/src/features/model-pricing/pricing-currency-selector.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-inputs.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/stores/pricing-preferences-store.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-99974a814f00

- Title: feat(plugins): extend plugin metadata and icon support
- Feature/group: feat(plugins): extend plugin metadata and icon support
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_plugin.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/icon.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/request.go`, `plugins/embed.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/google/plugin.js`, `plugins/tasks/hailuo/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `plugins/tasks/vertex-ai/plugin.js`, `plugins/tasks/vidu/plugin.js`, `router/api-router.go`, `router/plugin_router_test.go`
- Overlap with downstream changes: controller/channel.go, router/api-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-984330920061

- Title: feat(web): improve plugin management and marketplace
- Feature/group: feat(web): improve plugin management and marketplace
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-plugin-logo.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace.test.ts`, `web/src/features/task-plugins/__tests__/plugin-card.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon-file.test.ts`, `web/src/features/task-plugins/__tests__/plugin-icon-image.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon.test.ts`, `web/src/features/task-plugins/__tests__/plugin-meta-preview.test.ts`, `web/src/features/task-plugins/__tests__/plugins-table.test.tsx`, `web/src/features/task-plugins/__tests__/website.test.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/javascript-viewer.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-card.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-endpoints.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-integrity-check.tsx`, `web/src/features/task-plugins/components/plugin-metadata-card.tsx`, `web/src/features/task-plugins/components/plugin-model-list.tsx`, `web/src/features/task-plugins/components/plugin-website-link.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/components/source-diff.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/lib/marketplace.ts`, `web/src/features/task-plugins/lib/plugin-icon-file.ts`, `web/src/features/task-plugins/lib/plugin-icon.ts`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`, `web/src/features/task-plugins/lib/plugin-website.ts`, `web/src/features/task-plugins/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/test-setup.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-eb76b136b85e

- Title: feat(channels): improve plugin channel setup and icons
- Feature/group: feat(channels): improve plugin channel setup and icons
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/provider-badge.tsx`, `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/channel-type-badge.test.tsx`, `web/src/features/channels/components/channel-type-badge.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/__tests__/task-plugin-base-url.test.ts`, `web/src/features/channels/lib/task-plugin-base-url.ts`, `web/src/features/channels/types.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/channels-table.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0e0ba152bdcc

- Title: feat(pricing): improve pricing editors and log display
- Feature/group: feat(pricing): improve pricing editors and log display
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-dialogs.tsx`, `web/src/features/models/components/models-provider.tsx`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/components/settings-form-layout.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/model-pricing-inputs.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`
- Overlap with downstream changes: web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/usage-logs/components/columns/common-logs-columns.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-71c1fd7caad7

- Title: feat(models): improve model listing, pricing and visibility filters
- Feature/group: feat(models): improve model listing, pricing and visibility filters
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_meta.go`, `model/model_meta.go`, `model/pricing.go`, `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/models/__tests__/metadata-editing.test.tsx`, `web/src/features/models/__tests__/model-listing.test.tsx`, `web/src/features/models/components/data-table-bulk-actions.tsx`, `web/src/features/models/components/data-table-row-actions.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/model-connections.tsx`, `web/src/features/models/components/model-square-status.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/lib/model-utils.ts`, `web/src/features/models/types.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/price.ts`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/routes/_authenticated/models/$section.tsx`
- Overlap with downstream changes: web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a5e41a893379

- Title: feat(usage-logs): refine mobile layout and keep quick actions visible
- Feature/group: feat(usage-logs): refine mobile layout and keep quick actions visible
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/core/__tests__/pagination.test.tsx`, `web/src/components/data-table/core/pagination.tsx`, `web/src/components/data-table/layout/data-table-page.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/components/usage-logs-mobile-card.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/compact-date-time-range-picker.tsx, web/src/features/usage-logs/components/usage-logs-mobile-card.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2bec370629aa

- Title: feat(web): refine API key and user quota displays
- Feature/group: feat(web): refine API key and user quota displays
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/core/data-table-view.tsx`, `web/src/components/data-table/core/truncated-cell.tsx`, `web/src/components/group-badge.tsx`, `web/src/components/ui/table.tsx`, `web/src/features/keys/components/__tests__/api-key-group-cell.test.tsx`, `web/src/features/keys/components/__tests__/api-key-group-combobox.test.tsx`, `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/__tests__/auto-group-order-editor.test.tsx`, `web/src/features/keys/components/api-key-group-cell.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-key-timestamp-cell.tsx`, `web/src/features/keys/components/api-keys-cells.tsx`, `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/keys/components/api-keys-table.tsx`, `web/src/features/keys/components/auto-group-visuals.tsx`, `web/src/features/keys/components/data-table-row-actions.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/user-quota-cell.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/styles/index.css`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/styles/index.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8f72ecbbfd86

- Title: feat(usage-logs): add searchable group filter
- Feature/group: feat(usage-logs): add searchable group filter
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/usage-logs/components/__tests__/group-filter.test.tsx`, `web/src/features/usage-logs/components/__tests__/log-type-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-551bb63edf40

- Title: fix(keys): show desktop quota amounts side by side
- Feature/group: fix(keys): show desktop quota amounts side by side
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bd22e45a740a

- Title: style(keys): widen the desktop quota column
- Feature/group: style(keys): widen the desktop quota column
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-keys-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-950644c9d544

- Title: fix(keys): preserve spacing after desktop quota content
- Feature/group: fix(keys): preserve spacing after desktop quota content
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-keys-table.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-524455fac3c4

- Title: feat(redemptions): add batch deletion and optional file exports
- Feature/group: feat(redemptions): add batch deletion and optional file exports
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/redemption.go`, `controller/redemption_batch_test.go`, `middleware/audit.go`, `model/redemption.go`, `router/api-router.go`, `web/src/features/redemption-codes/api.ts`, `web/src/features/redemption-codes/components/__tests__/bulk-delete.test.tsx`, `web/src/features/redemption-codes/components/__tests__/redemptions-export.test.tsx`, `web/src/features/redemption-codes/components/data-table-bulk-actions.tsx`, `web/src/features/redemption-codes/components/redemptions-export-dialog.tsx`, `web/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx`, `web/src/features/redemption-codes/components/redemptions-table.tsx`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/audit.go, router/api-router.go, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-fff0635bb14b

- Title: docs: update project architecture and Go conventions
- Feature/group: docs: update project architecture and Go conventions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ebe4c368f287

- Title: refactor: modernize Go code conventions
- Feature/group: refactor: modernize Go code conventions
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/custom-event.go`, `common/etag.go`, `common/gopool.go`, `common/init.go`, `common/quota_math.go`, `common/redis.go`, `common/session_cookie.go`, `common/ssrf_protection.go`, `common/str.go`, `common/totp.go`, `common/utils.go`, `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/auth_flow_test.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_authz_test.go`, `controller/channel_upstream_update.go`, `controller/channel_upstream_update_test.go`, `controller/custom_oauth.go`, `controller/deployment.go`, `controller/email_binding.go`, `controller/group.go`, `controller/model.go`, `controller/model_management_test.go`, `controller/model_meta.go`, `controller/model_pricing_config.go`, `controller/model_sync.go`, `controller/oauth.go`, `controller/option.go`, `controller/passkey.go`, `controller/pricing.go`, `controller/ratio_sync.go`, `controller/redemption.go`, `controller/relay_error_log_test.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/subscription.go`, `controller/task_plugin.go`, `controller/telegram_test.go`, `controller/topup.go`, `controller/topup_waffo.go`, `controller/topup_waffo_pancake_test.go`, `controller/twofa.go`, `controller/uptime_kuma.go`, `controller/user.go`, `controller/wechat.go`, `dto/midjourney.go`, `dto/suno.go`, `middleware/audit.go`, `middleware/auth.go`, `middleware/rate-limit.go`, `middleware/task_artifact_access.go`, `model/audit_log.go`, `model/channel.go`, `model/channel_cache.go`, `model/channel_satisfy.go`, `model/channel_status_test.go`, `model/checkin.go`, `model/gorm_logger.go`, `model/json_column_test.go`, `model/log.go`, `model/log_format_test.go`, `model/log_other_test.go`, `model/main.go`, `model/model_pricing_config.go`, `model/option.go`, `model/passkey.go`, `model/perf_metric.go`, `model/prefill_group.go`, `model/pricing.go`, `model/quota_reserve.go`, `model/quota_reserve_test.go`, `model/redemption.go`, `model/redemption_test.go`, `model/subscription.go`, `model/task.go`, `model/task_cas_test.go`, `model/token.go`, `model/topup.go`, `model/twofa.go`, `model/usedata.go`, `model/user.go`, `model/user_auth_cache.go`, `model/user_authentication_test.go`, `model/user_cache.go`, `model/user_session.go`, `model/user_session_test.go`, `model/user_update_test.go`, `model/utils.go`, `oauth/registry.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/run.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `plugins/hailuo_responses_test.go`, `relay/channel/ali/dto.go`, `relay/channel/api_request_getbody_test.go`, `relay/channel/aws/dto.go`, `relay/channel/aws/relay-aws.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/gemini/adaptor.go`, `relay/channel/gemini/adaptor_responses_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/ollama/dto.go`, `relay/channel/ollama/relay-ollama.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/chat_via_responses_test.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_image.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/tencent/dto.go`, `relay/channel/tencent/relay-tencent.go`, `relay/common/outbound_body_test.go`, `relay/common/override.go`, `relay/common/override_test.go`, `relay/common/relay_info.go`, `relay/common/relay_utils.go`, `relay/common/stream_status_test.go`, `relay/common/tool_usage_test.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/common.go`, `relay/helper/stream_scanner_test.go`, `relay/mjproxy_handler.go`, `relay/plugin_protocol.go`, `relay/relay_task.go`, `relay/request_clone_test.go`, `relaykit/dto/billing_usage.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/notify.go`, `relaykit/dto/openai_image.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/pricing.go`, `relaykit/dto/ratio_sync.go`, `relaykit/dto/realtime.go`, `relaykit/relayconvert/internal/claude_messages/citations.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/req_helpers.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_stream_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_stream_resp.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/kitutil/mask.go`, `relaykit/relayconvert/kitutil/value.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/response_registry_test.go`, `router/plugin-router.go`, `service/auth_session.go`, `service/auth_session_test.go`, `service/auth_token.go`, `service/authz/adapter.go`, `service/authz/registry.go`, `service/authz/resolver.go`, `service/channel_affinity.go`, `service/channel_affinity_template_test.go`, `service/email_binding.go`, `service/error_test.go`, `service/file_decoder.go`, `service/file_service.go`, `service/funding_source.go`, `service/group.go`, `service/http_client_transport_test.go`, `service/http_transport_sharded.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/notify-limit.go`, `service/quota.go`, `service/quota_saturation_test.go`, `service/rankings.go`, `service/system_instance.go`, `service/task_billing_test.go`, `service/task_plugin_audit.go`, `service/task_polling.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/token_counter.go`, `service/violation_fee.go`, `service/webhook.go`, `setting/auto_group.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `setting/config/config.go`, `setting/model_setting/claude.go`, `setting/model_setting/gemini.go`, `setting/model_setting/global.go`, `setting/operation_setting/channel_affinity_setting.go`, `setting/operation_setting/operation_setting.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `setting/ratio_setting/exposed_cache.go`, `setting/ratio_setting/model_ratio.go`, `setting/sensitive.go`, `setting/system_setting/task_artifact_store.go`, `setting/task_plugin.go`, `setting/user_usable_group.go`
- Overlap with downstream changes: common/redis.go, common/ssrf_protection.go, controller/audit.go, controller/channel-test.go, controller/channel.go, controller/channel_authz_test.go, controller/model.go, controller/option.go, controller/secure_verification.go, controller/user.go, model/channel.go, model/channel_cache.go, model/gorm_logger.go, model/log.go, model/log_format_test.go, model/main.go, model/option.go, model/task.go, model/task_cas_test.go, model/twofa.go, model/usedata.go, model/user.go, model/user_authentication_test.go, model/user_cache.go, model/user_update_test.go, relay/channel/openai/adaptor.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relay/helper/common.go, relay/mjproxy_handler.go, relay/relay_task.go, relaykit/dto/openai_request.go, relaykit/relayconvert/kitutil/mask.go, service/auth_session.go, service/auth_session_test.go, service/auth_token.go, service/funding_source.go, service/http_client_transport_test.go, service/midjourney.go, service/quota.go, service/task_billing_test.go, service/task_polling.go, service/text_quota.go, service/violation_fee.go, service/webhook.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ea7cb0ba4e0f

- Title: refactor(web): unify table cells and quota details
- Feature/group: refactor(web): unify table cells and quota details
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/activity-time-cell.tsx`, `web/src/components/data-table/core/data-table-view.tsx`, `web/src/components/group-badge.tsx`, `web/src/components/masked-value-display.tsx`, `web/src/components/quota-details-popover.tsx`, `web/src/components/ui/table.tsx`, `web/src/features/keys/components/__tests__/api-key-group-cell.test.tsx`, `web/src/features/keys/components/__tests__/api-key-group-combobox.test.tsx`, `web/src/features/keys/components/__tests__/api-key-listing.test.tsx`, `web/src/features/keys/components/api-key-quota-cell.tsx`, `web/src/features/keys/components/api-key-timestamp-cell.tsx`, `web/src/features/keys/components/api-keys-cells.tsx`, `web/src/features/redemption-codes/components/redemptions-columns.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/user-quota-cell.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/users/components/users-table.tsx`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9bf328d97497

- Title: fix: preserve provider fields in Sora video queries
- Feature/group: fix: preserve provider fields in Sora video queries
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/README.md`, `docs/plugin-api/v1.md`, `plugins/tasks/sora/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4fc9d1f1fa77

- Title: fix(options): rebuild options table primary key and stop pricing writes resetting rows
- Feature/group: fix(options): rebuild options table primary key and stop pricing writes resetting rows
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/main.go`, `model/model_pricing_config.go`, `model/option_primary_key_migration.go`
- Overlap with downstream changes: model/main.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c79b74b68358

- Title: fix(frontend): deduplicate /api/status requests (#7189)
- Feature/group: fix(frontend): deduplicate /api/status requests (#7189)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/hooks/use-status.ts`, `web/src/hooks/use-system-config.ts`, `web/src/lib/__tests__/status-query.test.tsx`, `web/src/lib/nav-modules.ts`, `web/src/lib/status-query.ts`, `web/src/main.tsx`, `web/src/routes/pricing/$modelId/index.tsx`, `web/src/routes/pricing/index.tsx`, `web/src/routes/rankings/index.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/hooks/use-system-config.ts, web/src/lib/nav-modules.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-876903a8eb22

- Title: fix: 修正火山方舟渠道获取模型列表的端点路径 (#7203)
- Feature/group: fix: 修正火山方舟渠道获取模型列表的端点路径 (#7203)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel_upstream_update.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d52bdc0b4087

- Title: feat(billing): add time-based pricing editor and expression previews
- Feature/group: feat(billing): add time-based pricing editor and expression previews
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/components/pricing-columns.tsx`, `web/src/features/pricing/hooks/use-billing-time.ts`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/condition-display.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/structure.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/billing-time-fields.tsx`, `web/src/features/system-settings/models/draft-number-input.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/system-settings/models/visual-condition-tree.css`, `web/src/features/system-settings/models/visual-condition-tree.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-12be9975c0bf

- Title: fix(web): unify server error notifications
- Feature/group: fix(web): unify server error notifications
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/AGENTS.md`, `web/src/components/layout/components/chat-presets-item.tsx`, `web/src/components/sign-out-dialog.tsx`, `web/src/features/about/index.tsx`, `web/src/features/auth/forgot-password/components/forgot-password-form.tsx`, `web/src/features/auth/hooks/use-email-verification.ts`, `web/src/features/auth/hooks/use-oauth-login.ts`, `web/src/features/auth/otp/components/otp-form.tsx`, `web/src/features/auth/reset-password-confirm/index.tsx`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/auth/sign-up/components/sign-up-form.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/channel-type-badge.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/channels/components/dialogs/balance-query-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/codex-usage-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/fetch-models-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/ollama-models-dialog.tsx`, `web/src/features/channels/components/dialogs/tag-batch-edit-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/use-channel-key-disclosure.ts`, `web/src/features/channels/hooks/use-channel-mutate-form.ts`, `web/src/features/channels/hooks/use-channel-upstream-updates.ts`, `web/src/features/channels/index.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/chat/hooks/use-active-chat-key.ts`, `web/src/features/dashboard/components/flow/flow-charts.tsx`, `web/src/features/dashboard/components/models/performance-overview.tsx`, `web/src/features/dashboard/components/overview/overview-dashboard.tsx`, `web/src/features/dashboard/components/overview/performance-health-panel.tsx`, `web/src/features/dashboard/components/overview/summary-cards.tsx`, `web/src/features/dashboard/components/users/user-charts.tsx`, `web/src/features/dashboard/lib/flow-selection.ts`, `web/src/features/home/hooks/use-home-page-content.ts`, `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/keys/components/api-keys-delete-dialog.tsx`, `web/src/features/keys/components/api-keys-multi-delete-dialog.tsx`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/keys/components/api-keys-provider.tsx`, `web/src/features/keys/components/api-keys-table.tsx`, `web/src/features/keys/components/data-table-row-actions.tsx`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`, `web/src/features/model-pricing/__tests__/save-errors.test.tsx`, `web/src/features/model-pricing/api.ts`, `web/src/features/models/components/deployments-table.tsx`, `web/src/features/models/components/dialogs/create-deployment-drawer.tsx`, `web/src/features/models/components/dialogs/extend-deployment-dialog.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/dialogs/prefill-group-management-dialog.tsx`, `web/src/features/models/components/dialogs/rename-deployment-dialog.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/features/models/components/dialogs/update-config-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-operation-dialog.tsx`, `web/src/features/models/components/dialogs/vendors-management-dialog.tsx`, `web/src/features/models/components/dialogs/view-details-dialog.tsx`, `web/src/features/models/components/dialogs/view-logs-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/models/components/drawers/prefill-group-form-drawer.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/models/components/vendor-linked-models.tsx`, `web/src/features/models/components/vendors-table.tsx`, `web/src/features/models/index.tsx`, `web/src/features/models/lib/model-actions.ts`, `web/src/features/models/lib/vendor-actions.ts`, `web/src/features/models/vendor-api.ts`, `web/src/features/playground/api.ts`, `web/src/features/playground/hooks/use-chat-handler.ts`, `web/src/features/playground/hooks/use-playground-options.ts`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-details-performance.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/hooks/use-pricing-data.ts`, `web/src/features/profile/components/checkin-calendar-card.tsx`, `web/src/features/profile/components/language-preferences-card.tsx`, `web/src/features/profile/components/sidebar-modules-card.tsx`, `web/src/features/profile/components/tabs/notification-tab.tsx`, `web/src/features/profile/hooks/use-profile.ts`, `web/src/features/rankings/hooks/use-rankings.ts`, `web/src/features/redemption-codes/components/__tests__/bulk-delete.test.tsx`, `web/src/features/redemption-codes/components/__tests__/redemptions-mutate-drawer.test.tsx`, `web/src/features/redemption-codes/components/data-table-bulk-actions.tsx`, `web/src/features/redemption-codes/components/data-table-row-actions.tsx`, `web/src/features/redemption-codes/components/redemptions-delete-dialog.tsx`, `web/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx`, `web/src/features/redemption-codes/components/redemptions-primary-buttons.tsx`, `web/src/features/redemption-codes/components/redemptions-table.tsx`, `web/src/features/security/components/__tests__/privacy-card.test.tsx`, `web/src/features/security/components/dialogs/telegram-bind-dialog.tsx`, `web/src/features/security/components/login-sessions-card.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/security/components/privacy-card.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-account-security.ts`, `web/src/features/security/hooks/use-two-fa-setup.ts`, `web/src/features/setup/setup-wizard.tsx`, `web/src/features/subscriptions/components/dialogs/reset-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/subscription-purchase-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/toggle-status-dialog.tsx`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/subscriptions/components/subscriptions-table.tsx`, `web/src/features/system-info/components/system-instances-panel.tsx`, `web/src/features/system-info/components/system-tasks-panel.tsx`, `web/src/features/system-settings/auth/custom-oauth/hooks/use-custom-oauth-mutations.ts`, `web/src/features/system-settings/auth/custom-oauth/hooks/use-custom-oauth-providers.ts`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/features/system-settings/content/announcements-section.tsx`, `web/src/features/system-settings/content/api-info-section.tsx`, `web/src/features/system-settings/content/faq-section.tsx`, `web/src/features/system-settings/content/uptime-kuma-section.tsx`, `web/src/features/system-settings/general/channel-affinity/cache-stats-dialog.tsx`, `web/src/features/system-settings/general/channel-affinity/index.tsx`, `web/src/features/system-settings/hooks/use-system-options.ts`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/integrations/payment-settings-section.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/system-settings/maintenance/log-settings-section.tsx`, `web/src/features/system-settings/maintenance/performance-section.tsx`, `web/src/features/system-settings/maintenance/update-checker-section.tsx`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/upstream-ratio-sync.tsx`, `web/src/features/task-plugins/api.ts`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/components/marketplace-plugin-card.tsx`, `web/src/features/task-plugins/components/marketplace-sources-dialog.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/plugin-icon.tsx`, `web/src/features/task-plugins/components/plugin-sandbox.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`, `web/src/features/task-plugins/index.tsx`, `web/src/features/usage-logs/audit/api.ts`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/common-logs-stats.tsx`, `web/src/features/usage-logs/components/dialogs/user-info-dialog.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/users/api.ts`, `web/src/features/users/components/data-table-row-actions.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/users/components/user-quota-dialog.tsx`, `web/src/features/users/components/users-delete-dialog.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/features/wallet/components/subscription-plans-card.tsx`, `web/src/features/wallet/hooks/use-affiliate.ts`, `web/src/features/wallet/hooks/use-billing-history.ts`, `web/src/features/wallet/hooks/use-creem-payment.ts`, `web/src/features/wallet/hooks/use-payment.ts`, `web/src/features/wallet/hooks/use-redemption.ts`, `web/src/features/wallet/hooks/use-topup-info.ts`, `web/src/features/wallet/hooks/use-waffo-pancake-payment.ts`, `web/src/features/wallet/hooks/use-waffo-payment.ts`, `web/src/hooks/use-notifications.ts`, `web/src/lib/__tests__/server-error-notifications.test.ts`, `web/src/lib/api.ts`, `web/src/lib/handle-server-error.ts`, `web/src/lib/http-client.ts`, `web/src/lib/query-client.ts`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.ts`, `web/src/main.tsx`, `web/src/routes/(auth)/oauth.tsx`, `web/src/routes/_authenticated/chat2link.tsx`, `web/src/routes/oauth/$provider.tsx`
- Overlap with downstream changes: web/src/features/about/index.tsx, web/src/features/auth/otp/components/otp-form.tsx, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/channels/api.ts, web/src/features/channels/components/channels-table.tsx, web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/hooks/use-channel-mutate-form.ts, web/src/features/dashboard/components/overview/overview-dashboard.tsx, web/src/features/dashboard/components/overview/summary-cards.tsx, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/features/system-info/components/system-tasks-panel.tsx, web/src/features/system-settings/content/announcements-section.tsx, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/maintenance/log-settings-section.tsx, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/features/usage-logs/components/common-logs-stats.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx, web/src/features/users/api.ts, web/src/features/users/components/data-table-row-actions.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/users/components/users-mutate-drawer.tsx, web/src/features/users/components/users-table.tsx, web/src/lib/http-client.ts, web/src/lib/server-error-message.ts, web/src/routes/oauth/$provider.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a20574136b27

- Title: fix(alibaba): correct Wan model protocols and usage accounting
- Feature/group: fix(alibaba): correct Wan model protocols and usage accounting
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/alibaba_wan3_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7cf9b473f61e

- Title: fix(web): keep model pricing content in one scroll area
- Feature/group: fix(web): keep model pricing content in one scroll area
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-064ed943e1ac

- Title: feat(billing): support fixed per-request expression pricing
- Feature/group: feat(billing): support fixed per-request expression pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/fixed.go`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `pkg/billingexpr/types.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `relay/relay_task.go`, `service/log_info_generate.go`, `service/quota.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `setting/billing_setting/tiered_billing.go`, `setting/billing_setting/tiered_billing_test.go`, `web/src/features/pricing/__tests__/breakdown-tier-match.test.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/relay_task.go, service/quota.go, service/text_quota.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bdef11750524

- Title: fix: restore add split button in advanced custom routes (#7289)
- Feature/group: fix: restore add split button in advanced custom routes (#7289)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b6566f33d908

- Title: feat: clarify passkey website configuration
- Feature/group: feat: clarify passkey website configuration
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/system-settings/hooks/use-update-option.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3e1b8b153249

- Title: feat(plugins): display localized changelogs with English fallback
- Feature/group: feat(plugins): display localized changelogs with English fallback
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`, `web/src/components/ui/markdown.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-changelog-panel.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-changelog.test.ts`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/plugin-changelog-panel.tsx`, `web/src/features/task-plugins/lib/plugin-changelog.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3cea2bf799a6

- Title: feat(relaykit): preserve cached input token breakdowns
- Feature/group: feat(relaykit): preserve cached input token breakdowns
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/dto/billing_usage.go`, `relaykit/dto/billing_usage_test.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/usage_merge.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-f362c7c51b86

- Title: feat(billingexpr): support image cache and quantity variables
- Feature/group: feat(billingexpr): support image cache and quantity variables
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/billingexpr/billingexpr_test.go`, `pkg/billingexpr/compile.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/fixed.go`, `pkg/billingexpr/run.go`, `pkg/billingexpr/settle.go`, `pkg/billingexpr/testdata/frontend_simulation.json`, `pkg/billingexpr/types.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-f064bffa2b65

- Title: fix(billing): validate image quantities before reserving quota
- Feature/group: fix(billing): validate image quantities before reserving quota
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/model.go`, `common/model_test.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/dto.go`, `relay/channel/ali/image.go`, `relay/channel/ali/image_wan.go`, `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`, `relay/common/relay_info.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/openai_image_request_test.go`, `relay/helper/price.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relaykit/dto/legacy_dalle_image.go`, `relaykit/dto/openai_image.go`, `service/billing_session.go`, `service/image_billing.go`, `service/log_info_generate.go`, `service/quota.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `service/tiered_settle_test.go`
- Overlap with downstream changes: relay/common/relay_info.go, service/billing_session.go, service/quota.go, service/text_quota.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-f256e40bcb29

- Title: feat(billing): add expression defaults for GPT image models
- Feature/group: feat(billing): add expression defaults for GPT image models
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `setting/billing_setting/builtin_billing.go`, `setting/billing_setting/builtin_billing_test.go`, `setting/billing_setting/tiered_billing.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-25ec832faa48

- Title: feat(pricing): convert legacy prices into expression drafts
- Feature/group: feat(pricing): convert legacy prices into expression drafts
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/model_management_test.go`, `controller/model_pricing_config.go`, `model/legacy_dalle_pricing.go`, `model/model_pricing_config.go`, `model/model_pricing_conversion.go`, `router/api-router.go`, `setting/ratio_setting/cache_ratio.go`, `setting/ratio_setting/model_ratio.go`
- Overlap with downstream changes: router/api-router.go, setting/ratio_setting/model_ratio.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-39294418af4c

- Title: feat(web): extend expression pricing editors and usage details
- Feature/group: feat(web): extend expression pricing editors and usage details
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/parser.ts`, `web/src/features/pricing/lib/billing-expression/runtime.ts`, `web/src/features/pricing/lib/billing-expression/types.ts`, `web/src/features/pricing/lib/billing-expression/visual.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-expr.ts`, `web/src/features/pricing/lib/tier-expr.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/request-simulation.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/models/tier-price-fields.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/features/system-settings/models/visual-billing-document-editor.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b3e279464c82

- Title: feat(web): review legacy pricing conversion before applying drafts
- Feature/group: feat(web): review legacy pricing conversion before applying drafts
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/model-pricing/__tests__/editor-currency.test.tsx`, `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing-conversion-dialog.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-505805a4c3ab

- Title: feat(web): unify channel setup and refine model and settings editors
- Feature/group: feat(web): unify channel setup and refine model and settings editors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/selection.test.tsx`, `web/src/components/ui/__tests__/label.test.tsx`, `web/src/components/ui/label.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/dialogs/fetch-models-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/hooks/use-channel-model-discovery.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`, `web/src/features/models/components/dialogs/vendor-mutate-dialog.tsx`, `web/src/features/models/components/drawers/model-mutate-drawer.tsx`, `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/general/channel-affinity/rule-editor-dialog.tsx`, `web/src/features/system-settings/integrations/waffo-settings-section.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/models/components/drawers/model-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d4c26bfb8dcb

- Title: fix(channels): allow editing multi-key selection strategy
- Feature/group: fix(channels): allow editing multi-key selection strategy
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/hooks/use-channel-mutate-form.ts`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/hooks/use-channel-mutate-form.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ab489ab8842b

- Title: feat(channels): add plugin extensions to provider and model selection
- Feature/group: feat(channels): add plugin extensions to provider and model selection
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/components/channel-plugin-extensions.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-plugin-extensions.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-251b76d8633e

- Title: feat(web): add administrator update reminders
- Feature/group: feat(web): add administrator update reminders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/layout/components/app-header.tsx`, `web/src/components/layout/components/public-header.tsx`, `web/src/components/layout/components/system-brand.tsx`, `web/src/features/system-settings/maintenance/update-checker-section.tsx`, `web/src/features/system-update/__tests__/releases.test.ts`, `web/src/features/system-update/__tests__/update-checking.test.tsx`, `web/src/features/system-update/api.ts`, `web/src/features/system-update/releases.ts`, `web/src/features/system-update/store.ts`, `web/src/features/system-update/system-update-action.tsx`, `web/src/features/system-update/system-update-dialog.tsx`, `web/src/features/system-update/use-system-update.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/components/layout/components/public-header.tsx, web/src/components/layout/components/system-brand.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-74629e29f83f

- Title: feat(plugins): enhance task streaming and model pricing
- Feature/group: feat(plugins): enhance task streaming and model pricing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `controller/billing_option_test.go`, `controller/model_management_test.go`, `controller/option.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/channel_constraints.go`, `dto/task.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `model/channel_constraint.go`, `model/model_pricing_config.go`, `model/pricing.go`, `model/pricing_usage_schema_test.go`, `model/task_plugin_channel_select_test.go`, `pkg/billingexpr/expr.md`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/alibaba_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/performance_test.go`, `relay/channel/task/jsplugin/submit_stream.go`, `relay/helper/stream_scanner.go`, `relay/helper/stream_scanner_test.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `service/channel_select.go`, `service/channel_select_test.go`, `service/task_billing.go`, `service/task_polling.go`, `setting/billing_setting/tiered_billing.go`, `web/src/features/model-pricing/__tests__/editor-layout.test.tsx`, `web/src/features/model-pricing/__tests__/pricing.test.ts`, `web/src/features/model-pricing/api.ts`, `web/src/features/model-pricing/model-pricing-panel.tsx`, `web/src/features/model-pricing/pricing.ts`, `web/src/features/pricing/__tests__/dynamic-price.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/cached-price-cell.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-price-cell.tsx`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/plugin-pricing.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/pricing/types.ts`, `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/models/__tests__/plugin-pricing.test.tsx`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`, `web/src/features/system-settings/models/index.tsx`, `web/src/features/system-settings/models/model-pricing-core.ts`, `web/src/features/system-settings/models/model-pricing-sheet.tsx`, `web/src/features/system-settings/models/model-pricing-snapshots.ts`, `web/src/features/system-settings/models/model-ratio-form.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/ratio-settings-card.tsx`, `web/src/features/system-settings/models/task-plugin-pricing-editor.tsx`, `web/src/features/system-settings/models/task-pricing-matrix.tsx`, `web/src/features/system-settings/models/task-usage-pricing-editor.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/__tests__/usage-schema-table.test.tsx`, `web/src/features/task-plugins/components/plugin-detail-sheet.tsx`, `web/src/features/task-plugins/components/usage-schema-table.tsx`, `web/src/features/task-plugins/types.ts`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: AGENTS.md, controller/option.go, controller/relay.go, dto/task.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go, relay/relay_task.go, service/task_billing.go, service/task_polling.go, web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/models/index.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-385d2dfd10d8

- Title: feat(auth): add safe multi-RP ID passkey support
- Feature/group: feat(auth): add safe multi-RP ID passkey support
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/audit.go`, `controller/login_verification.go`, `controller/misc.go`, `controller/option.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `model/option.go`, `model/passkey.go`, `model/passkey_option.go`, `model/user_authentication_test.go`, `router/api-router.go`, `service/account_security.go`, `service/passkey/service.go`, `service/passkey/session.go`, `service/security_verification.go`, `setting/system_setting/passkey.go`, `web/src/features/auth/passkey/__tests__/domain-verification.test.tsx`, `web/src/features/auth/passkey/api.ts`, `web/src/features/auth/passkey/assertion.ts`, `web/src/features/auth/passkey/components/passkey-domain-selector.tsx`, `web/src/features/auth/passkey/types.ts`, `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/api.ts`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/auth/sign-in/components/user-auth-form.tsx`, `web/src/features/system-settings/api.ts`, `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/index.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/lib/server-error-message.ts`
- Overlap with downstream changes: controller/audit.go, controller/misc.go, controller/option.go, controller/secure_verification.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/option.go, model/user_authentication_test.go, router/api-router.go, web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx, web/src/features/auth/secure-verification/types.ts, web/src/features/auth/sign-in/components/user-auth-form.tsx, web/src/features/system-settings/api.ts, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/lib/server-error-message.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-129f21b6942e

- Title: fix(plugins): normalize invalid UTF-8 in JSON state
- Feature/group: fix(plugins): normalize invalid UTF-8 in JSON state
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c9a110190c52

- Title: fix(channels): show built-in base URLs as placeholders
- Feature/group: fix(channels): show built-in base URLs as placeholders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `constant/channel.go`, `controller/channel.go`, `controller/channel_test_internal_test.go`, `router/channel-router.go`, `router/channel_router_test.go`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/channels/lib/channel-type-config.ts`
- Overlap with downstream changes: constant/channel.go, controller/channel.go, router/channel-router.go, router/channel_router_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-type-config.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-be36cbb8facf

- Title: test(security): provide query client for passkey verification tests
- Feature/group: test(security): provide query client for passkey verification tests
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/__tests__/enrollment.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-007d69942d24

- Title: fix(model): migrate renamed prefill group unique indexes
- Feature/group: fix(model): migrate renamed prefill group unique indexes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/prefill_group_migration.go`, `model/prefill_group_migration_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2ba61576146f

- Title: feat(web): share collapsible mobile table filters
- Feature/group: feat(web): share collapsible mobile table filters
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/data-table/index.ts`, `web/src/components/data-table/toolbar/__tests__/mobile-filter.test.tsx`, `web/src/components/data-table/toolbar/mobile-filter-panel.tsx`, `web/src/components/data-table/toolbar/toolbar.tsx`, `web/src/features/channels/components/channels-table.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`
- Overlap with downstream changes: web/src/features/channels/components/channels-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-043ff99a51ec

- Title: fix: handle legacy database constraints and scoped policies
- Feature/group: fix: handle legacy database constraints and scoped policies
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `model/migration_dialector.go`, `model/migration_dialector_test.go`, `service/authz/adapter.go`, `service/authz/authz_test.go`, `service/authz/seed.go`
- Overlap with downstream changes: service/authz/authz_test.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-33142f0aee7c

- Title: fix: refine sign-in page behavior
- Feature/group: fix: refine sign-in page behavior
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/misc.go`, `controller/passkey_test.go`, `web/src/features/auth/passkey/__tests__/domain-verification.test.tsx`, `web/src/features/auth/passkey/components/passkey-domain-selector.tsx`
- Overlap with downstream changes: controller/misc.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7fd06381976c

- Title: fix(auth): restore saved compatible passkey domains
- Feature/group: fix(auth): restore saved compatible passkey domains
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/passkey-settings.test.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/auth/section-registry.tsx`
- Overlap with downstream changes: web/src/features/system-settings/auth/section-registry.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-815217ba648f

- Title: test(web): stabilize flaky frontend suite timeouts and motion visibility races (#7367)
- Feature/group: test(web): stabilize flaky frontend suite timeouts and motion visibility races (#7367)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/test-setup.ts`, `web/vitest.config.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d92612038410

- Title: fix: unify model provider detection and add Wan icon (#7373)
- Feature/group: fix: unify model provider detection and add Wan icon (#7373)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/assets/custom/icon-wan.tsx`, `web/src/assets/custom/wan.png`, `web/src/features/channels/lib/model-categories.ts`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/lib/lobe-icon.tsx`, `web/src/lib/model-provider.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d1c79d728822

- Title: fix(dashboard): align weekly default range (#7355)
- Feature/group: fix(dashboard): align weekly default range (#7355)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/constants.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2509e25fa060

- Title: fix(web): prevent combobox dropdowns opening on dialog autofocus (#7365)
- Feature/group: fix(web): prevent combobox dropdowns opening on dialog autofocus (#7365)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/__tests__/combobox.test.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-76f7dafd2b82

- Title: fix(audio): normalize file extension case in GetAudioDuration (#7321)
- Feature/group: fix(audio): normalize file extension case in GetAudioDuration (#7321)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/audio.go`, `common/audio_case_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-04c64734cc7c

- Title: fix(pricing): keep model status bar spacing uniform (#7284)
- Feature/group: fix(pricing): keep model status bar spacing uniform (#7284)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9fe0457ee1f4

- Title: Add Responses WebSocket relay support (#5062)
- Feature/group: Add Responses WebSocket relay support (#5062)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/json.go`, `controller/channel.go`, `controller/channel_multi_key_test.go`, `controller/relay.go`, `controller/responses_websocket.go`, `controller/responses_websocket_test.go`, `dto/channel_constraints.go`, `main.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/task_plugin_origin_task_test.go`, `model/channel.go`, `model/channel_constraint.go`, `pkg/wsmanager/wsmanager.go`, `pkg/wsmanager/wsmanager_test.go`, `relay/channel/openai/relay_responses.go`, `relay/chat_completions_via_responses_test.go`, `relay/helper/valid_request.go`, `relay/mjproxy_handler.go`, `relay/relay_task_test.go`, `relay/request_billing.go`, `relay/responses_handler.go`, `relay/responses_request.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`, `relay/websocket.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp_test.go`, `router/relay-router.go`, `service/channel.go`, `service/relay_error.go`, `service/relay_error_test.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/ws_close.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/responses-websocket-setting.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/responses-websocket-setting.tsx`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/channel.go, controller/relay.go, main.go, model/channel.go, relay/chat_completions_via_responses_test.go, relay/mjproxy_handler.go, relay/responses_handler.go, relaykit/dto/channel_settings.go, router/relay-router.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8529f209c859

- Title: feat: vllm channel && sglang channel (#7332)
- Feature/group: feat: vllm channel && sglang channel (#7332)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.gitignore`, `common/advanced_custom_presets.go`, `common/api_type.go`, `common/endpoint_type.go`, `constant/channel.go`, `controller/channel.go`, `controller/channel_inference.go`, `controller/channel_inference_test.go`, `controller/channel_upstream_update.go`, `model/channel.go`, `model/channel_constraint.go`, `model/channel_settings_test.go`, `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/presets_test.go`, `relay/channel/advancedcustom/rerank.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/relay_responses_billing_test.go`, `relay/common/relay_info.go`, `relay/helper/max_tokens_bounds_test.go`, `relay/helper/valid_request.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/claude.go`, `relaykit/dto/openai_compaction.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_request_zero_value_test.go`, `relaykit/dto/openai_response.go`, `relaykit/dto/values.go`, `relaykit/dto/values_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_resp.go`, `router/channel-router.go`, `router/channel_router_test.go`, `web/src/assets/brand-icons/sglang.svg`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/inference-status.test.tsx`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/channels-dialogs.tsx`, `web/src/features/channels/components/channels-provider.tsx`, `web/src/features/channels/components/dialogs/inference-status-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/inference-status.test.ts`, `web/src/features/channels/lib/__tests__/new-api-channel.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/lib/channel-type-config.ts`, `web/src/features/channels/lib/channel-utils.ts`, `web/src/features/channels/lib/inference-status.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: .gitignore, common/api_type.go, constant/channel.go, controller/channel.go, model/channel.go, model/channel_settings_test.go, relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/openai_request.go, relaykit/dto/openai_request_zero_value_test.go, relaykit/dto/values.go, relaykit/dto/values_test.go, router/channel-router.go, router/channel_router_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/channels-provider.tsx, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/lib/channel-type-config.ts, web/src/features/channels/lib/channel-utils.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2bfb89c1b99d

- Title: fix(pricing): prevent sort menu layout shift (#7145)
- Feature/group: fix(pricing): prevent sort menu layout shift (#7145)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/components/pricing-toolbar.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8e5e09166cf8

- Title: fix(ollama): preserve tool calls from final stream frame | 保留流式末尾帧的 tool calls (#7376)
- Feature/group: fix(ollama): preserve tool calls from final stream frame | 保留流式末尾帧的 tool calls (#7376)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/stream.go`, `relay/channel/ollama/stream_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-62f8db775bcd

- Title: feat(channel): 修复获取DeepSeek余额错误 (#6814)
- Feature/group: feat(channel): 修复获取DeepSeek余额错误 (#6814)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel-billing.go`, `controller/channel_billing_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d3874db61fce

- Title: feat(rate-limit): optimize in-memory limiter allocation and cleanup (#6807)
- Feature/group: feat(rate-limit): optimize in-memory limiter allocation and cleanup (#6807)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/rate-limit.go`, `common/rate-limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-5caafd3d84dc

- Title: feat(ollama): add per-channel OpenAI-compatible chat switch (#7382)
- Feature/group: feat(ollama): add per-channel OpenAI-compatible chat switch (#7382)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/ollama/adaptor.go`, `relaykit/dto/channel_settings.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relaykit/dto/channel_settings.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-fa90b2312cfb

- Title: fix(gemini): accept case-insensitive thinkingLevel and log canonical effort (#7387)
- Feature/group: fix(gemini): accept case-insensitive thinkingLevel and log canonical effort (#7387)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-81336fc69b8a

- Title: fix(gemini): reject :countTokens as an unknown route instead of relaying it as generateContent (#7388)
- Feature/group: fix(gemini): reject :countTokens as an unknown route instead of relaying it as generateContent (#7388)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `router/relay-router.go`
- Overlap with downstream changes: router/relay-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-610334dbd760

- Title: fix(relay): request stream usage on every cross-protocol conversion to OpenAI chat (#7389)
- Feature/group: fix(relay): request stream usage on every cross-protocol conversion to OpenAI chat (#7389)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/adaptor_test.go`, `relay/channel/openai/adaptor.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-69a50029819a

- Title: feat(channel): per-route pass-through for advanced custom channels (#7386)
- Feature/group: feat(channel): per-route pass-through for advanced custom channels (#7386)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/advancedcustom/adaptor.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/advanced-custom.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7209b6db95ef

- Title: fix(middleware): 隐藏无可用渠道错误中的任务插件标识 (#7414)
- Feature/group: fix(middleware): 隐藏无可用渠道错误中的任务插件标识 (#7414)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/distributor.go`, `middleware/distributor_test.go`
- Overlap with downstream changes: i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, middleware/distributor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-57791e31c62b

- Title: docs(github): tighten issue and PR templates for pass-through and unfiltered AI text (#7438)
- Feature/group: docs(github): tighten issue and PR templates for pass-through and unfiltered AI text (#7438)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.agents/github/ISSUE.md`, `.agents/github/PR.md`, `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE/en.md`, `AGENTS.md`
- Overlap with downstream changes: .github/PULL_REQUEST_TEMPLATE.md, AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ed0ad6e3ff9f

- Title: feat(channels): add compact quick options to channel editor
- Feature/group: feat(channels): add compact quick options to channel editor
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-07578faf26d1

- Title: fix(billing): count upstream images by payload instead of data length
- Feature/group: fix(billing): count upstream images by payload instead of data length
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d82678a9708b

- Title: docs: move billing rules to .agents/rules/billing.md behind a read gate
- Feature/group: docs: move billing rules to .agents/rules/billing.md behind a read gate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `AGENTS.md`, `CLAUDE.md`
- Overlap with downstream changes: AGENTS.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0cde9d94f6d3

- Title: feat(channels): add floating model redirect workbench to channel editor
- Feature/group: feat(channels): add floating model redirect workbench to channel editor
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/floating-window.tsx`, `web/src/components/floating-window/__tests__/floating-window.test.tsx`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/option-hint.test.tsx`, `web/src/components/ui/combobox-input.tsx`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-batch-dialog.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/dialogs/configure-models-dialog.tsx`, `web/src/features/channels/components/drawers/channel-configuration.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/model-mapping-batch-dialog.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/model-redirect-panel.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/features/channels/hooks/__tests__/use-redirect-panel-placement.test.tsx`, `web/src/features/channels/hooks/use-redirect-panel-placement.ts`, `web/src/features/channels/lib/__tests__/model-mapping-rules.test.ts`, `web/src/features/channels/lib/__tests__/model-naming-patterns.test.ts`, `web/src/features/channels/lib/index.ts`, `web/src/features/channels/lib/model-mapping-rules.ts`, `web/src/features/channels/lib/model-naming-patterns.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8b2c7105f91f

- Title: feat(relay): classify the protocol outcome on StreamStatus
- Feature/group: feat(relay): classify the protocol outcome on StreamStatus
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/common/stream_status.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-73a471f3a8b1

- Title: feat(policy): add request policies settings and routing decision records
- Feature/group: feat(policy): add request policies settings and routing decision records
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_pin_retry_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/request_policy.go`, `middleware/distributor.go`, `model/option.go`, `model/request_policy.go`, `model/request_policy_test.go`, `relay/mjproxy_handler.go`, `relay/request_billing.go`, `router/api-router.go`, `service/channel_affinity.go`, `service/channel_affinity_template_test.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/relay_error.go`, `service/relay_error_test.go`, `service/request_policy.go`, `setting/operation_setting/channel_affinity_setting.go`, `web/.oxlintrc.json`, `web/src/components/layout/config/system-settings.config.ts`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/inference-status.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/index.tsx`, `web/src/features/channels/types.ts`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/system-settings/components/settings-accordion.tsx`, `web/src/features/system-settings/components/settings-page.tsx`, `web/src/features/system-settings/general/channel-affinity/index.tsx`, `web/src/features/system-settings/general/channel-affinity/rule-editor-dialog.tsx`, `web/src/features/system-settings/general/channel-affinity/session-rules-table.tsx`, `web/src/features/system-settings/general/channel-affinity/types.ts`, `web/src/features/system-settings/models/index.tsx`, `web/src/features/system-settings/models/routing-reliability-section.tsx`, `web/src/features/system-settings/models/section-registry.tsx`, `web/src/features/system-settings/request-limits/sensitive-words-section.tsx`, `web/src/features/system-settings/request-policies/__tests__/routing.test.tsx`, `web/src/features/system-settings/request-policies/__tests__/settings.test.tsx`, `web/src/features/system-settings/request-policies/api.ts`, `web/src/features/system-settings/request-policies/channel-health-section.tsx`, `web/src/features/system-settings/request-policies/decision-record.tsx`, `web/src/features/system-settings/request-policies/defaults.ts`, `web/src/features/system-settings/request-policies/index.tsx`, `web/src/features/system-settings/request-policies/policy-label.ts`, `web/src/features/system-settings/request-policies/related-policy-link.tsx`, `web/src/features/system-settings/request-policies/request-checks-section.tsx`, `web/src/features/system-settings/request-policies/retry-section.tsx`, `web/src/features/system-settings/request-policies/routing-form.ts`, `web/src/features/system-settings/request-policies/routing-section.tsx`, `web/src/features/system-settings/request-policies/section-registry.tsx`, `web/src/features/system-settings/request-policies/use-save-policy.ts`, `web/src/features/system-settings/security/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/routeTree.gen.ts`, `web/src/routes/_authenticated/system-settings/models/$section.tsx`, `web/src/routes/_authenticated/system-settings/operations/$section.tsx`, `web/src/routes/_authenticated/system-settings/request-policies/$section.tsx`, `web/src/routes/_authenticated/system-settings/request-policies/index.tsx`, `web/src/routes/_authenticated/system-settings/security/$section.tsx`
- Overlap with downstream changes: controller/channel.go, controller/relay.go, middleware/distributor.go, model/option.go, relay/mjproxy_handler.go, router/api-router.go, service/midjourney.go, web/src/components/layout/config/system-settings.config.ts, web/src/features/channels/api.ts, web/src/features/channels/types.ts, web/src/features/system-settings/models/index.tsx, web/src/features/system-settings/models/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/routeTree.gen.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8a11b4305ed1

- Title: refactor(relaykit): surface reasoning conversion diagnostics
- Feature/group: refactor(relaykit): surface reasoning conversion diagnostics
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/adaptor.go`, `relay/convert_request_error_test.go`, `relay/helper/reasoning_suffix.go`, `relaykit/relayconvert/claude_default_max_tokens_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_responses_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req_test.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/gemini_test.go`, `relaykit/relayconvert/reasoning/intent.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_compat.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: relay/channel/openai/adaptor.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a5b40663d330

- Title: perf(metrics): record relay outcomes at the request boundary and classify stream results
- Feature/group: perf(metrics): record relay outcomes at the request boundary and classify stream results
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `constant/context_key.go`, `controller/perf_metrics.go`, `controller/relay.go`, `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `pkg/perf_metrics/outcome.go`, `pkg/perf_metrics/types.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/stream_tool_billing_test.go`, `relay/common/relay_info.go`, `relay/helper/stream_result.go`, `service/log_info_generate.go`, `service/quota.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/text_quota.go`, `web/src/features/dashboard/components/models/performance-overview.tsx`, `web/src/features/dashboard/components/overview/performance-health-panel.tsx`, `web/src/features/performance-metrics/__tests__/summary.test.tsx`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-details-performance.tsx`, `web/src/features/pricing/components/model-details-uptime-sparkline.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: constant/context_key.go, controller/relay.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, service/quota.go, service/text_quota.go, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6237d9d77eb0

- Title: fix(rate-limit): reserve model rate-limit slots and judge success by response outcome
- Feature/group: fix(rate-limit): reserve model rate-limit slots and judge success by response outcome
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/rate-limit.go`, `middleware/model-rate-limit.go`, `middleware/model_rate_limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-75f3d246e2b1

- Title: fix(responses-ws): carry stream_id through events and correlate errors
- Feature/group: fix(responses-ws): carry stream_id through events and correlate errors
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-90134d5b6bff

- Title: feat(system-tasks): add filtered task history with cleanup
- Feature/group: feat(system-tasks): add filtered task history with cleanup
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/system_task.go`, `controller/system_task_test.go`, `model/system_task.go`, `model/system_task_test.go`, `router/api-router.go`, `web/src/features/system-info/__tests__/system-tasks-panel.test.tsx`, `web/src/features/system-info/api.ts`, `web/src/features/system-info/components/system-task-history.tsx`, `web/src/features/system-info/components/system-tasks-panel.tsx`, `web/src/features/system-info/components/system-tasks-table.tsx`, `web/src/features/system-info/constants.ts`, `web/src/features/system-info/types.ts`, `web/src/features/system-settings/api.ts`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/system_task.go, model/system_task.go, router/api-router.go, web/src/features/system-info/api.ts, web/src/features/system-info/components/system-tasks-panel.tsx, web/src/features/system-info/types.ts, web/src/features/system-settings/api.ts, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a8ed7f7c5f24

- Title: fix(plugins/hailuo): bill only the dimensions each model family supports
- Feature/group: fix(plugins/hailuo): bill only the dimensions each model family supports
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e2471403c779

- Title: fix(pricing): adjust billing expression condition display
- Feature/group: fix(pricing): adjust billing expression condition display
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/lib/__tests__/billing-expression.test.ts`, `web/src/features/pricing/lib/billing-expression/condition-display.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-42cbe6206a62

- Title: chore(web): add intl-locale lint plugin and update agent rules
- Feature/group: chore(web): add intl-locale lint plugin and update agent rules
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `web/.oxlintrc.json`, `web/AGENTS.md`, `web/scripts/oxlint/__tests__/intl-locale.test.ts`, `web/scripts/oxlint/intl-locale.mjs`, `web/src/features/system-settings/models/__tests__/visual-billing-editor.test.tsx`, `web/src/features/system-settings/models/billing-time-fields.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/tsconfig.node.json`, `web/vitest.config.ts`
- Overlap with downstream changes: AGENTS.md, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d8069436612e

- Title: fix(web): tidy channel health policy layout
- Feature/group: fix(web): tidy channel health policy layout
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/request-policies/__tests__/channel-health-layout.test.tsx`, `web/src/features/system-settings/request-policies/channel-health-section.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ae249f4ecbaf

- Title: feat(responses-ws): extend channel support and share routing with HTTP
- Feature/group: feat(responses-ws): extend channel support and share routing with HTTP
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/relay.go`, `controller/responses_websocket_test.go`, `middleware/distributor.go`, `middleware/distributor_test.go`, `model/channel_constraint.go`, `relay/channel/api_request.go`, `relay/channel/api_request_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`, `relaykit/dto/channel_settings.go`, `service/channel_select.go`
- Overlap with downstream changes: controller/relay.go, middleware/distributor.go, relaykit/dto/channel_settings.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-fa3cc1c6b9b1

- Title: feat(web): offer the Responses WebSocket toggle for advanced custom, sub2api and new-api channels
- Feature/group: feat(web): offer the Responses WebSocket toggle for advanced custom, sub2api and new-api channels
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/drawers/channel-provider-picker.tsx`, `web/src/features/channels/components/responses-websocket-setting.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/responses-websocket.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/constants.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-32ef6d216f9f

- Title: feat(web): add channel passthrough controls and usage guidance
- Feature/group: feat(web): add channel passthrough controls and usage guidance
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/channel-quick-options.test.tsx`, `web/src/features/channels/components/channel-quick-options.tsx`, `web/src/features/channels/components/dialogs/passthrough-warning-dialog.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/__tests__/header-passthrough.test.ts`, `web/src/features/channels/lib/header-passthrough.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9e905a83e1d6

- Title: refactor: show subscription deductions and improve cost display (#7272)
- Feature/group: refactor: show subscription deductions and improve cost display (#7272)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/columns.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ed7c4e35d429

- Title: feat(log): flag upstream response model mismatches (#7418)
- Feature/group: feat(log): flag upstream response model mismatches (#7418)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/common/relay_info.go`, `relay/common/response_model.go`, `relay/response_model_test.go`, `service/log_info_generate.go`, `service/relay_error.go`, `service/responses_usage.go`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e2bb05f70ed6

- Title: fix(relay): map Images output_tokens_details into img_o billing (#7410)
- Feature/group: fix(relay): map Images output_tokens_details into img_o billing (#7410)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/image_stream_test.go`, `relay/channel/openai/relay_image.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-55a6cd2a44ea

- Title: fix(responses): estimate usage for streams cut before terminal usage (#7427)
- Feature/group: fix(responses): estimate usage for streams cut before terminal usage (#7427)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `service/responses_usage.go`, `service/responses_usage_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3524fe0b1579

- Title: fix issue template (#7441)
- Feature/group: fix issue template (#7441)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/bug_report_en.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml`, `.github/ISSUE_TEMPLATE/feature_request_en.yml`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2906e4f779b7

- Title: fix(oauth): require account evidence before migrating legacy GitHub bindings
- Feature/group: fix(oauth): require account evidence before migrating legacy GitHub bindings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/auth_flow_test.go`, `controller/auth_session_test.go`, `controller/login_verification.go`, `controller/model_list_test.go`, `controller/oauth.go`, `controller/passkey_test.go`, `controller/user.go`, `controller/wechat.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `model/account_security.go`, `model/login_verification.go`, `model/user.go`, `oauth/github.go`, `oauth/provider.go`, `service/login_verification.go`
- Overlap with downstream changes: controller/model_list_test.go, controller/user.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/user.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3e8c358da09b

- Title: fix: forward reasoning_effort to zhipu upstream (#7124)
- Feature/group: fix: forward reasoning_effort to zhipu upstream (#7124)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/zhipu_4v/relay-zhipu_v4.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-946c730ed008

- Title: perf: avoid repeated Claude tool-name lookups (#7447)
- Feature/group: perf: avoid repeated Claude tool-name lookups (#7447)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/dto/claude.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ea336ed6654a

- Title: fix(log): hide funding source labels when subscriptions are not in play (#7450)
- Feature/group: fix(log): hide funding source labels when subscriptions are not in play (#7450)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/__tests__/billing-source.test.ts`, `web/src/features/usage-logs/lib/billing-source.ts`, `web/src/features/usage-logs/lib/columns.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-96b0da2273a1

- Title: refactor(log): 恢复费用气泡并用图标区分订阅来源 (#7451)
- Feature/group: refactor(log): 恢复费用气泡并用图标区分订阅来源 (#7451)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/billing-source-visibility.test.tsx`, `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/usage-logs/lib/__tests__/billing-source.test.ts`, `web/src/features/usage-logs/lib/billing-source.ts`
- Overlap with downstream changes: web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1ea24a050d5e

- Title: docs: refresh project README across languages
- Feature/group: docs: refresh project README across languages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e04c05c2b929

- Title: fix(web): 修复小数输入限制和日志分组选项未遮挡的问题 (#7453)
- Feature/group: fix(web): 修复小数输入限制和日志分组选项未遮挡的问题 (#7453)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ui/combobox-input.tsx`, `web/src/components/ui/combobox.tsx`, `web/src/features/system-settings/models/__tests__/group-pricing-decimals.test.tsx`, `web/src/features/system-settings/models/__tests__/tool-price-validation.test.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/tool-price-settings.tsx`, `web/src/features/usage-logs/components/__tests__/group-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/common-logs-filter-bar.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b0bf2580e249

- Title: docs: restore README badge and partner placement
- Feature/group: docs: restore README badge and partner placement
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-972aed197282

- Title: fix(log): derive response model mismatch from names instead of a stored flag (#7464)
- Feature/group: fix(log): derive response model mismatch from names instead of a stored flag (#7464)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/common/response_model.go`, `relay/response_model_test.go`, `service/log_info_generate.go`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/__tests__/usage-facts.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`, `web/src/features/usage-logs/lib/__tests__/response-model.test.ts`, `web/src/features/usage-logs/lib/index.ts`, `web/src/features/usage-logs/lib/response-model.ts`, `web/src/features/usage-logs/types.ts`
- Overlap with downstream changes: web/src/features/usage-logs/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a89037d0b9b0

- Title: fix(web): start linked-account verification from the provider button
- Feature/group: fix(web): start linked-account verification from the provider button
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/auth/secure-verification/__tests__/verification-flow.test.tsx`, `web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx`, `web/src/features/auth/secure-verification/hooks/use-secure-verification.ts`
- Overlap with downstream changes: web/src/features/auth/secure-verification/components/secure-verification-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ef70d1b79998

- Title: chore: drop trailing blank line from .gitignore
- Feature/group: chore: drop trailing blank line from .gitignore
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.gitignore`
- Overlap with downstream changes: .gitignore
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b932e6d876cd

- Title: fix(web): robustly display conditional task billing prices
- Feature/group: fix(web): robustly display conditional task billing prices
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-price-display.test.tsx`, `web/src/features/pricing/components/dynamic-pricing-breakdown.tsx`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/pricing/lib/billing-expr.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/task-display.ts`, `web/src/features/pricing/lib/breakdown-tier-match.ts`, `web/src/features/pricing/lib/dynamic-price.ts`, `web/src/features/pricing/lib/task-matrix-display.ts`, `web/src/features/pricing/lib/task-price-display.ts`, `web/src/features/system-settings/models/__tests__/request-simulation.test.tsx`, `web/src/features/system-settings/models/tiered-pricing-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9c3d3aeb3c30

- Title: perf(web): lazy-load provider icons and reduce page rerenders
- Feature/group: perf(web): lazy-load provider icons and reduce page rerenders
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-keys-columns.tsx`, `web/src/features/models/components/models-columns.tsx`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/__tests__/pricing-controls.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-card.tsx`, `web/src/features/pricing/components/pricing-sidebar.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/drawing-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/usage-logs-table.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/lib/__tests__/lobe-icon.test.tsx`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/usage-logs-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-33320407f4f5

- Title: fix(web): isolate OIDC discovery when saving OAuth settings
- Feature/group: fix(web): isolate OIDC discovery when saving OAuth settings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/auth/__tests__/oauth-settings.test.tsx`, `web/src/features/system-settings/auth/oauth-section.tsx`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4c34f25a49f1

- Title: feat(channel): bind multiple task plugins to New API channels
- Feature/group: feat(channel): bind multiple task plugins to New API channels
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel.go`, `controller/channel_task_plugin_bind_test.go`, `controller/channel_task_plugin_validation_test.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `middleware/distributor.go`, `middleware/distributor_test.go`, `middleware/task_plugin.go`, `model/channel_constraint.go`, `model/task.go`, `model/task_plugin.go`, `model/task_plugin_channel_select_test.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `plugins/builtin_plugins_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth_test.go`, `relaykit/dto/channel_settings.go`, `service/task_polling.go`, `service/task_polling_test.go`, `web/src/components/multi-select.tsx`, `web/src/components/multi-select/__tests__/option-icon.test.tsx`, `web/src/features/channels/api.ts`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/__tests__/new-api-channel.test.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/lib/channel-plugin-extensions.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/channel.go, middleware/distributor.go, model/task.go, relaykit/dto/channel_settings.go, service/task_polling.go, service/task_polling_test.go, web/src/features/channels/api.ts, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2d7aef7414e8

- Title: fix(web): restore CC Switch model dropdown
- Feature/group: fix(web): restore CC Switch model dropdown
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/dialogs/__tests__/cc-switch-dialog.test.tsx`, `web/src/features/keys/components/dialogs/cc-switch-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6b638788c645

- Title: fix(channels): preview upstream model changes before applying
- Feature/group: fix(channels): preview upstream model changes before applying
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/upstream-update-dialog.test.tsx`, `web/src/features/channels/components/channels-dialogs.tsx`, `web/src/features/channels/components/data-table-row-actions.tsx`, `web/src/features/channels/components/dialogs/upstream-update-dialog.tsx`, `web/src/features/channels/hooks/use-channel-upstream-updates.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-db536ec8bcbb

- Title: fix(channels): simplify model redirects and clarify panel controls
- Feature/group: fix(channels): simplify model redirects and clarify panel controls
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/floating-window.tsx`, `web/src/components/floating-window/__tests__/floating-window.test.tsx`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-batch-dialog.test.tsx`, `web/src/features/channels/components/__tests__/upstream-model-selection.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/components/model-mapping-batch-dialog.tsx`, `web/src/features/channels/components/model-mapping-editor.tsx`, `web/src/features/channels/components/model-redirect-panel.tsx`, `web/src/features/channels/components/upstream-model-selection.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-58a9eff4c842

- Title: style: refresh logo
- Feature/group: style: refresh logo
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/icon.png`, `electron/tray-icon-windows.png`, `electron/tray-iconTemplate.png`, `electron/tray-iconTemplate@2x.png`, `web/index.html`, `web/public/favicon.ico`, `web/public/logo.png`, `web/rsbuild.config.ts`
- Overlap with downstream changes: web/index.html, web/public/favicon.ico, web/public/logo.png, web/rsbuild.config.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3fdf9083dc6b

- Title: fix(settings): move documentation link to site settings
- Feature/group: fix(settings): move documentation link to site settings
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/general/quota-settings-section.tsx`, `web/src/features/system-settings/general/system-info-section.tsx`, `web/src/features/system-settings/site/index.tsx`, `web/src/features/system-settings/site/section-registry.tsx`, `web/src/features/system-settings/types.ts`
- Overlap with downstream changes: web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/general/system-info-section.tsx, web/src/features/system-settings/site/index.tsx, web/src/features/system-settings/site/section-registry.tsx, web/src/features/system-settings/types.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-da2540dda132

- Title: fix(channels): align card metrics and remove badge spacing
- Feature/group: fix(channels): align card metrics and remove badge spacing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/channel-card.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d904b9700015

- Title: feat(settings): improve group configuration UX and drag sorting
- Feature/group: feat(settings): improve group configuration UX and drag sorting
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/auto-group-order-item.tsx`, `web/src/features/keys/components/auto-group-order-editor.tsx`, `web/src/features/system-settings/models/__tests__/group-settings.test.tsx`, `web/src/features/system-settings/models/group-ratio-form.tsx`, `web/src/features/system-settings/models/group-ratio-visual-editor.tsx`, `web/src/features/system-settings/models/group-special-usable-editor.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-585c4610e5d7

- Title: feat(keys): show and copy configured API addresses
- Feature/group: feat(keys): show and copy configured API addresses
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/__tests__/api-addresses.test.tsx`, `web/src/features/keys/components/api-keys-primary-buttons.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c49972441e73

- Title: fix(web): keep mobile dialog actions within visible viewport
- Feature/group: fix(web): keep mobile dialog actions within visible viewport
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/dialog.tsx`, `web/src/components/ui/__tests__/dialog-layout.test.tsx`, `web/src/components/ui/alert-dialog.tsx`, `web/src/components/ui/dialog.tsx`, `web/src/features/channels/components/dialogs/advanced-custom-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/channel-test-dialog.tsx`, `web/src/features/channels/components/dialogs/edit-tag-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`, `web/src/features/channels/components/dialogs/upstream-update-dialog.tsx`, `web/src/features/model-pricing/pricing-conversion-dialog.tsx`, `web/src/features/models/components/dialogs/missing-models-dialog.tsx`, `web/src/features/models/components/dialogs/update-config-dialog.tsx`, `web/src/features/models/components/dialogs/view-details-dialog.tsx`, `web/src/features/system-settings/auth/custom-oauth/components/provider-form-dialog.tsx`, `web/src/features/system-settings/models/channel-selector-dialog.tsx`, `web/src/features/system-settings/models/conflict-confirm-dialog.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/wallet/components/dialogs/billing-history-dialog.tsx`, `web/src/styles/index.css`
- Overlap with downstream changes: web/src/features/channels/components/dialogs/channel-test-dialog.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/styles/index.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-03563a4a7916

- Title: feat(plugins): serve OpenAI Images API through task plugins
- Feature/group: feat(plugins): serve OpenAI Images API through task plugins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `common/model.go`, `constant/channel.go`, `controller/model_management_test.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_image.go`, `controller/plugin_protocol_test.go`, `controller/relay.go`, `controller/relay_task_plugin_test.go`, `controller/task.go`, `controller/task_generic_test.go`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `dto/task.go`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `model/model_pricing_conversion.go`, `model/task.go`, `pkg/billingexpr/expr.md`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `plugins/alibaba_responses_test.go`, `plugins/builtin_plugins_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `relay/channel/ali/adaptor.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/dto.go`, `relay/channel/ali/image.go`, `relay/channel/ali/image_wan.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/convert_request_error_test.go`, `relay/helper/billing_expr_request.go`, `relay/helper/openai_image_request_test.go`, `relay/helper/price.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `router/relay-router.go`, `router/task-plugin-protocol-router.go`, `router/task_plugin_protocol_router_test.go`, `service/image_billing.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/text_quota_test.go`, `setting/model_setting/qwen.go`, `web/public/favicon.ico`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/__tests__/provider-selection.test.tsx`, `web/src/features/channels/constants.ts`, `web/src/features/channels/lib/channel-utils.ts`, `web/src/features/task-plugins/__tests__/plugin-detail-sheet.test.tsx`, `web/src/features/task-plugins/lib/host-protocols.ts`, `web/src/features/usage-logs/__tests__/artifacts.test.ts`, `web/src/features/usage-logs/components/__tests__/billing-source-visibility.test.tsx`, `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/__tests__/task-artifacts-cell.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/audio-preview-dialog.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`, `web/src/features/usage-logs/components/task-artifacts.tsx`, `web/src/features/usage-logs/components/timing-metrics-cell.tsx`, `web/src/features/usage-logs/lib/task-artifacts.ts`, `web/src/features/usage-logs/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: constant/channel.go, controller/relay.go, controller/task.go, dto/task.go, model/task.go, relay/relay_task.go, router/relay-router.go, service/task_billing.go, service/task_billing_test.go, web/public/favicon.ico, web/src/features/channels/constants.ts, web/src/features/channels/lib/channel-utils.ts, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a5bf94022060

- Title: fix(web): correct dialog viewport fallbacks and editor gutter alignment
- Feature/group: fix(web): correct dialog viewport fallbacks and editor gutter alignment
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ai-elements/__tests__/code-block-editor.test.tsx`, `web/src/components/ai-elements/code-block.tsx`, `web/src/components/dialog.tsx`, `web/src/components/ui/__tests__/dialog-layout.test.tsx`, `web/src/components/ui/alert-dialog.tsx`, `web/src/components/ui/dialog.tsx`, `web/src/features/system-settings/models/__tests__/group-pricing-decimals.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2fddb863f221

- Title: feat(models): add bulk field selection to metadata sync
- Feature/group: feat(models): add bulk field selection to metadata sync
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/models/__tests__/metadata-sync.test.tsx`, `web/src/features/models/components/dialogs/sync-wizard-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3abbb8198e04

- Title: feat: record perf metrics when async tasks reach terminal status (#7481)
- Feature/group: feat: record perf metrics when async tasks reach terminal status (#7481)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `service/task_polling.go`
- Overlap with downstream changes: service/task_polling.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-65d3a2171fd1

- Title: fix(plugins): price only the resolutions each Seedance and Wan model offers
- Feature/group: fix(plugins): price only the resolutions each Seedance and Wan model offers
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9978ee1e25a6

- Title: feat(billing): configure trust threshold and input pre-consume multiplier
- Feature/group: feat(billing): configure trust threshold and input pre-consume multiplier
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/constants.go`, `common/quota.go`, `controller/billing_option_test.go`, `model/option.go`, `pkg/billingexpr/expr.md`, `pkg/billingexpr/types.go`, `relay/helper/price.go`, `relay/helper/price_test.go`, `service/billing_session.go`, `service/image_billing.go`, `setting/operation_setting/quota_setting.go`, `web/src/features/system-settings/billing/index.tsx`, `web/src/features/system-settings/billing/section-registry.tsx`, `web/src/features/system-settings/general/__tests__/pre-consume-settings.test.tsx`, `web/src/features/system-settings/general/quota-settings-section.tsx`, `web/src/features/system-settings/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: common/constants.go, model/option.go, service/billing_session.go, web/src/features/system-settings/billing/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-47713bcb1749

- Title: feat: update plugins
- Feature/group: feat: update plugins
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/tasks/jimeng/plugin.js`, `plugins/tasks/kling/plugin.js`, `plugins/tasks/sora/plugin.js`, `plugins/tasks/sunoapi/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9a0be8750a6d

- Title: fix(web): keep task pricing editable after a plugin narrows its enum values
- Feature/group: fix(web): keep task pricing editable after a plugin narrows its enum values
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/task-expr.test.ts`, `web/src/features/pricing/__tests__/task-matrix-display.test.ts`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/lib/billing-expression/README.md`, `web/src/features/pricing/lib/billing-expression/display.ts`, `web/src/features/pricing/lib/billing-expression/task-display.ts`, `web/src/features/system-settings/models/__tests__/task-pricing-copy.test.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9c293e8c0237

- Title: test(billing): fix websocket rejection refund test after pre-consume dropped output estimation (#7491)
- Feature/group: test(billing): fix websocket rejection refund test after pre-consume dropped output estimation (#7491)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-474ed66fb14b

- Title: fix(plugins/alibaba): keep upstream metadata in openai_image render (#7507)
- Feature/group: fix(plugins/alibaba): keep upstream metadata in openai_image render (#7507)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/alibaba_responses_test.go`, `plugins/tasks/alibaba/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-00e8a00cb4ae

- Title: fix(web): include provider icons in Windows builds
- Feature/group: fix(web): include provider icons in Windows builds
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/lib/__tests__/lobe-icon.test.tsx`, `web/src/lib/lobe-icon.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e537dc380bb2

- Title: fix(web): expose plugin activation after upload
- Feature/group: fix(web): expose plugin activation after upload
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c0cff23a3af8

- Title: fix(plugins): store task plugin source as longtext on MySQL and raise the upload cap to 8 MiB
- Feature/group: fix(plugins): store task plugin source as longtext on MySQL and raise the upload cap to 8 MiB
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/task_plugin.go`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/marketplace-install-dialog.tsx`, `web/src/features/task-plugins/components/plugin-source-picker.tsx`, `web/src/features/task-plugins/components/plugin-url-import-field.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/lib/plugin-url.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2c175190c44a

- Title: perf(plugins): sync task plugins from hashes and load source only for changed rows
- Feature/group: perf(plugins): sync task plugins from hashes and load source only for changed rows
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_plugin.go`, `controller/task_plugin_test.go`, `model/task_plugin.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d61d6be7593f

- Title: fix(task): accept any 2xx upstream status on task submission
- Feature/group: fix(task): accept any 2xx upstream status on task submission
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `relay/relay_task.go`, `relay/relay_task_test.go`
- Overlap with downstream changes: relay/relay_task.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-54eee488bed3

- Title: fix(web): isolate theme preferences in local storage
- Feature/group: fix(web): isolate theme preferences in local storage
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/context/__tests__/theme-preferences.test.tsx`, `web/src/context/theme-customization-provider.tsx`, `web/src/context/theme-provider.tsx`, `web/src/lib/frontend-cache.ts`, `web/src/lib/theme-customization.ts`, `web/src/lib/theme-storage.ts`
- Overlap with downstream changes: web/src/lib/frontend-cache.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6e9de44a7dca

- Title: fix(relayconvert): hoist Responses tool-output media into a user message (#7510)
- Feature/group: fix(relayconvert): hoist Responses tool-output media into a user message (#7510)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4eb3b9160566

- Title: fix(relayconvert): hoist Claude tool_result images and segment Responses reasoning items (#7512)
- Feature/group: fix(relayconvert): hoist Claude tool_result images and segment Responses reasoning items (#7512)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b6809a52df88

- Title: fix(web): resolve safe constants in plugin metadata previews
- Feature/group: fix(web): resolve safe constants in plugin metadata previews
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-install-dialog.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-meta-preview.test.ts`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0aec08fee811

- Title: chore: credit #7444 author for the Claude tool_result fix landed in #7512
- Feature/group: chore: credit #7444 author for the Claude tool_result fix landed in #7512
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: (none)
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9310231b3c27

- Title: fix(logs): keep model badge text readable and compact (#7504)
- Feature/group: fix(logs): keep model badge text readable and compact (#7504)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/mobile-card.test.tsx`, `web/src/features/usage-logs/components/model-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8cb88ebd17a0

- Title: fix: move force operation into confirm slot so plugin-in-use dialog has one cancel button (#7528)
- Feature/group: fix: move force operation into confirm slot so plugin-in-use dialog has one cancel button (#7528)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/plugins-table.test.tsx`, `web/src/features/task-plugins/components/plugins-table.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-996adffe5165

- Title: fix(web): keep text selection when opening a model pricing row for editing (#7403)
- Feature/group: fix(web): keep text selection when opening a model pricing row for editing (#7403)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/__tests__/model-ratio-table-selection.test.tsx`, `web/src/features/system-settings/models/model-ratio-visual-editor.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-5401874c6387

- Title: fix(ci): align ldflags -X package path with module path so release binaries embed the real version (#7536)
- Feature/group: fix(ci): align ldflags -X package path with module path so release binaries embed the real version (#7536)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`
- Overlap with downstream changes: .github/workflows/electron-build.yml, .github/workflows/release.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-a46f045d6d14

- Title: fix: match current price column width to source columns so expressions wrap alike (#7542)
- Feature/group: fix: match current price column width to source columns so expressions wrap alike (#7542)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/upstream-ratio-sync-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6c14c0762dc7

- Title: feat: highlight differing values in source price expressions (#7543)
- Feature/group: feat: highlight differing values in source price expressions (#7543)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/system-settings/models/upstream-price-cells.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c76452d22412

- Title: fix(playground): add break-all to playground input (#6600)
- Feature/group: fix(playground): add break-all to playground input (#6600)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/ai-elements/prompt-input.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d04c118c8803

- Title: fix(channel): preserve Responses WebSocket setting on save (#7468)
- Feature/group: fix(channel): preserve Responses WebSocket setting on save (#7468)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/responses-websocket-setting.test.tsx`, `web/src/features/channels/lib/channel-form.ts`
- Overlap with downstream changes: web/src/features/channels/lib/channel-form.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c2b7a9a9e0b5

- Title: fix(claude): preserve per-message output_config in Claude messages (#7561)
- Feature/group: fix(claude): preserve per-message output_config in Claude messages (#7561)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/claude/adaptor_test.go`, `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-789c970199ea

- Title: fix(claude): preserve auto mode safeguards in native requests (#7598)
- Feature/group: fix(claude): preserve auto mode safeguards in native requests (#7598)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-f5e54ccdb203

- Title: fix(web): revalidate marketplace fetches so stale caches do not fail integrity checks
- Feature/group: fix(web): revalidate marketplace fetches so stale caches do not fail integrity checks
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/marketplace-panel.test.tsx`, `web/src/features/task-plugins/__tests__/marketplace-plugin-logo.test.tsx`, `web/src/features/task-plugins/__tests__/plugin-icon-file.test.ts`, `web/src/features/task-plugins/__tests__/plugin-url.test.ts`, `web/src/features/task-plugins/components/marketplace-panel.tsx`, `web/src/features/task-plugins/lib/plugin-icon-file.ts`, `web/src/features/task-plugins/lib/plugin-url.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-87bb71e7fc83

- Title: feat(plugins): serve Seedream images through the OpenAI Images protocol in doubao
- Feature/group: feat(plugins): serve Seedream images through the OpenAI Images protocol in doubao
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/doubao_responses_test.go`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3509ae7c194d

- Title: fix(pricing): keep prices saved before a usage profile narrows editable
- Feature/group: fix(pricing): keep prices saved before a usage profile narrows editable
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.md`, `model/model_pricing_config.go`, `model/pricing_usage_schema_test.go`, `pkg/billingexpr/expr.md`, `web/src/features/pricing/__tests__/task-matrix.test.ts`, `web/src/features/pricing/lib/billing-expression/display.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6cbb1c7edba4

- Title: feat(auth): require step-up verification for admin user management
- Feature/group: feat(auth): require step-up verification for admin user management
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/admin_user_verification_test.go`, `controller/custom_oauth.go`, `controller/passkey.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `docs/authentication.md`, `middleware/secure_verification.go`, `service/security_verification.go`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/users/api.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/features/users/components/__tests__/step-up-verification.test.tsx`, `web/src/features/users/components/data-table-row-actions.tsx`, `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/users/components/users-delete-dialog.tsx`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/features/users/components/users-provider.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, web/src/features/auth/secure-verification/types.ts, web/src/features/users/api.ts, web/src/features/users/components/data-table-row-actions.tsx, web/src/features/users/components/dialogs/user-binding-dialog.tsx, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-caca52f8dea0

- Title: feat(auth): replace the dashboard access token with scoped access tokens
- Feature/group: feat(auth): replace the dashboard access token with scoped access tokens
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/admin_user_verification_test.go`, `controller/audit.go`, `controller/auth_session.go`, `controller/custom_oauth.go`, `controller/email_binding.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/secure_verification.go`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `controller/twofa.go`, `controller/user.go`, `controller/user_manage_test.go`, `controller/wechat.go`, `docs/authentication.md`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `middleware/access_token_routes.go`, `middleware/audit.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/header_nav_test.go`, `middleware/secure_verification.go`, `model/audit_log.go`, `model/auth_flow.go`, `model/main.go`, `model/option.go`, `model/task_cas_test.go`, `model/user.go`, `model/user_access_token.go`, `model/user_authentication_test.go`, `model/user_update_test.go`, `router/access_token_scope_test.go`, `router/api-router.go`, `router/channel-router.go`, `router/plugin_router_test.go`, `service/access_token_scope.go`, `service/auth_cleanup.go`, `service/auth_session.go`, `service/auth_session_test.go`, `service/auth_token_test.go`, `service/security_verification.go`, `web/src/components/permission-matrix.tsx`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/__tests__/page.test.tsx`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-card.test.tsx`, `web/src/features/security/components/__tests__/access-token-create-dialog.test.tsx`, `web/src/features/security/components/__tests__/access-tokens-card.test.tsx`, `web/src/features/security/components/access-token-card.tsx`, `web/src/features/security/components/access-token-item.tsx`, `web/src/features/security/components/access-tokens-card.tsx`, `web/src/features/security/components/dialogs/access-token-create-dialog.tsx`, `web/src/features/security/components/dialogs/access-token-dialog.tsx`, `web/src/features/security/hooks/use-access-token.ts`, `web/src/features/security/hooks/use-access-tokens.ts`, `web/src/features/security/index.tsx`, `web/src/features/security/lib/access-token-schema.ts`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/__tests__/viewer.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/components/audit-log-filter-bar.tsx`, `web/src/features/usage-logs/audit/components/audit-log-viewer.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/users-mutate-drawer.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`, `web/src/lib/admin-permissions.ts`
- Overlap with downstream changes: controller/audit.go, controller/secure_verification.go, controller/user.go, controller/user_manage_test.go, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, model/auth_flow.go, model/main.go, model/option.go, model/task_cas_test.go, model/user.go, model/user_authentication_test.go, model/user_update_test.go, router/api-router.go, router/channel-router.go, service/auth_cleanup.go, service/auth_session.go, service/auth_session_test.go, web/src/features/auth/secure-verification/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/features/users/components/users-mutate-drawer.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts, web/src/lib/admin-permissions.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-335ebcff71b8

- Title: fix(relayconvert): keep Responses custom tools on Chat Completions upstreams
- Feature/group: fix(relayconvert): keep Responses custom tools on Chat Completions upstreams
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/chat_via_responses_test.go`, `relay/common/relay_info.go`, `relay/common/relay_info_test.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_registry.go`, `relaykit/relayconvert/response_registry_test.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2ffc59590115

- Title: feat(jsplugin): run task plugins on moejs
- Feature/group: feat(jsplugin): run task plugins on moejs
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `AGENTS.md`, `THIRD-PARTY-LICENSES.md`, `controller/plugin_protocol.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.md`, `dto/task_plugin.go`, `go.mod`, `go.sum`, `middleware/task_plugin.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `pkg/jsplugin/routing_test.go`, `pkg/jsplugin/utils.go`, `plugins/builtin_plugins_test.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/performance_test.go`, `relay/channel/task/jsplugin/submit_stream.go`, `relay/plugin_protocol.go`, `relay/plugin_protocol_test.go`, `service/task_plugin_view_test.go`
- Overlap with downstream changes: AGENTS.md, THIRD-PARTY-LICENSES.md, go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c71eefbcd194

- Title: docs: update readme
- Feature/group: docs: update readme
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `README.en.md`, `README.fr.md`, `README.ja.md`, `README.md`, `README.zh_CN.md`, `README.zh_TW.md`, `assets/screenshots/dashboard.en.jpg`, `assets/screenshots/dashboard.zh-CN.jpg`, `assets/screenshots/models.en.jpg`, `assets/screenshots/models.zh-CN.jpg`, `assets/screenshots/plugin-marketplace.en.jpg`, `assets/screenshots/plugin-marketplace.zh-CN.jpg`, `assets/screenshots/usage-logs.en.jpg`, `assets/screenshots/usage-logs.zh-CN.jpg`
- Overlap with downstream changes: README.md, README.zh_CN.md
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2506e1b980ec

- Title: fix(auth): reject non-standard roles when creating a user
- Feature/group: fix(auth): reject non-standard roles when creating a user
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/admin_user_verification_test.go`, `controller/user.go`
- Overlap with downstream changes: controller/user.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4924361aef73

- Title: feat(auth): edit the name and permissions of an access token
- Feature/group: feat(auth): edit the name and permissions of an access token
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/security_enrollment_test.go`, `model/user_access_token.go`, `model/user_update_test.go`, `router/api-router.go`, `service/access_token_scope.go`, `service/security_verification.go`, `web/src/components/permission-matrix.tsx`, `web/src/features/auth/secure-verification/types.ts`, `web/src/features/security/api.ts`, `web/src/features/security/components/__tests__/access-token-create-dialog.test.tsx`, `web/src/features/security/components/__tests__/access-tokens-card.test.tsx`, `web/src/features/security/components/access-token-item.tsx`, `web/src/features/security/components/access-tokens-card.tsx`, `web/src/features/security/components/dialogs/access-token-create-dialog.tsx`, `web/src/features/security/components/dialogs/access-token-edit-dialog.tsx`, `web/src/features/security/hooks/use-access-tokens.ts`, `web/src/features/security/lib/access-token-catalog.ts`, `web/src/features/security/lib/access-token-schema.ts`, `web/src/features/usage-logs/audit/__tests__/details.test.tsx`, `web/src/features/usage-logs/audit/components/audit-log-details-dialog.tsx`, `web/src/features/usage-logs/audit/lib/audit-details.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/users/components/__tests__/permissions.test.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/i18n/static-keys.ts`
- Overlap with downstream changes: controller/audit.go, model/user_update_test.go, router/api-router.go, web/src/features/auth/secure-verification/types.ts, web/src/features/usage-logs/lib/format.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/i18n/static-keys.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ae73ef8e25b9

- Title: feat(plugins): add Seedream 5.0 flash and price each Seedream model by its own tiers
- Feature/group: feat(plugins): add Seedream 5.0 flash and price each Seedream model by its own tiers
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/doubao_responses_test.go`, `plugins/tasks/doubao/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2035a82aeb54

- Title: feat(web): add a refresh button to the channels table
- Feature/group: feat(web): add a refresh button to the channels table
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/table-refresh.test.tsx`, `web/src/features/channels/components/channels-table.tsx`
- Overlap with downstream changes: web/src/features/channels/components/channels-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-feefe09f2781

- Title: fix: isolate TLS configs across HTTP transports (#7627)
- Feature/group: fix: isolate TLS configs across HTTP transports (#7627)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/init.go`, `controller/model_sync.go`, `controller/ratio_sync.go`, `service/http_client.go`, `service/protected_fetch_client.go`, `service/waffo_pancake.go`
- Overlap with downstream changes: service/http_client.go, service/protected_fetch_client.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d0cb7347c07b

- Title: fix: use max_completion_tokens for gpt-6-sol and gpt-6-luna (#7559)
- Feature/group: fix: use max_completion_tokens for gpt-6-sol and gpt-6-luna (#7559)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel_test_request_test.go`, `relaykit/dto/openai_request.go`
- Overlap with downstream changes: relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8112120673bf

- Title: fix(web): Add class to wrap the long strings at auto-disable status tooltip content in channel manage page (#7604)
- Feature/group: fix(web): Add class to wrap the long strings at auto-disable status tooltip content in channel manage page (#7604)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/channels/components/__tests__/channels-columns-status-tooltip-content.test.tsx`, `web/src/features/channels/components/channels-columns.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-dfd3cd893044

- Title: fix(web): hide system settings from non-root admins (#7513)
- Feature/group: fix(web): hide system settings from non-root admins (#7513)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/hooks/use-sidebar-data.ts`
- Overlap with downstream changes: web/src/hooks/use-sidebar-data.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-5fe8917be728

- Title: fix(i18n): keep cached zhTW/zhCN interface codes stable across page loads (#7139)
- Feature/group: fix(i18n): keep cached zhTW/zhCN interface codes stable across page loads (#7139)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/i18n/languages.test.ts`, `web/src/i18n/languages.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-56758edf95ec

- Title: fix(web): correct translations, chart ordering, and font loading (#7628)
- Feature/group: fix(web): correct translations, chart ordering, and font loading (#7628)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/dashboard/lib/__tests__/charts.test.ts`, `web/src/features/dashboard/lib/charts.ts`, `web/src/features/pricing/components/model-details.tsx`, `web/src/features/redemption-codes/components/redemptions-primary-buttons.tsx`, `web/src/features/security/__tests__/account-security.test.tsx`, `web/src/features/security/components/dialogs/delete-account-dialog.tsx`, `web/src/features/security/components/dialogs/two-fa-setup-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-columns.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/users/components/users-columns.tsx`, `web/src/features/wallet/__tests__/copy.test.tsx`, `web/src/features/wallet/components/dialogs/billing-history-dialog.tsx`, `web/src/features/wallet/components/recharge-form-card.tsx`, `web/src/features/wallet/lib/format.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/styles/theme.css`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/styles/theme.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1a4166d8e8ba

- Title: fix(relayconvert): keep Responses custom tools on Claude and Gemini upstreams (#7636)
- Feature/group: fix(relayconvert): keep Responses custom tools on Claude and Gemini upstreams (#7636)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/gemini/adaptor_responses_test.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp_test.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req_preprocess.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/policy_test.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/response_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b48b74ab77cf

- Title: feat(jsplugin): decode hook results with moejs ToGoInto
- Feature/group: feat(jsplugin): decode hook results with moejs ToGoInto
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `THIRD-PARTY-LICENSES.md`, `go.mod`, `go.sum`, `pkg/jsplugin/engine.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: THIRD-PARTY-LICENSES.md, go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-973cf8ef4600

- Title: fix(keys): prevent rounded clipping of group names (#7645)
- Feature/group: fix(keys): prevent rounded clipping of group names (#7645)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/keys/components/api-key-group-cell.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-78bd5b1cbf7b

- Title: fix(relaykit): Gemini stream token details, Responses tool-output media on Gemini, Opus 5.5 default effort (#7687)
- Feature/group: fix(relaykit): Gemini stream token details, Responses tool-output media on Gemini, Opus 5.5 default effort (#7687)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/dto/usage_merge.go`, `relaykit/dto/usage_merge_test.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/internal/shared/gemini/request.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-8ca26d81dfff

- Title: fix(web): stop frontend build files from exhausting the web rate limit
- Feature/group: fix(web): stop frontend build files from exhausting the web rate limit
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/constants.go`, `common/init.go`, `middleware/frontend_static.go`, `middleware/frontend_static_test.go`, `middleware/rate-limit.go`, `middleware/rate_limit_test.go`, `router/web-router.go`
- Overlap with downstream changes: common/constants.go, router/web-router.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-956019109bb0

- Title: feat(web): filter the user list by group
- Feature/group: feat(web): filter the user list by group
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/users/components/__tests__/group-filter.test.tsx`, `web/src/features/users/components/__tests__/quota-display.test.tsx`, `web/src/features/users/components/users-table.tsx`, `web/src/routes/_authenticated/users/index.tsx`
- Overlap with downstream changes: web/src/features/users/components/users-table.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c6bac617d46a

- Title: fix(relaykit): make the thinking suffix adapter toggle cover effort suffixes
- Feature/group: fix(relaykit): make the thinking suffix adapter toggle cover effort suffixes
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/helper/reasoning_suffix_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/reasoning/suffix.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `web/src/features/system-settings/models/claude-settings-card.tsx`, `web/src/features/system-settings/models/gemini-settings-card.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0908d9e7b911

- Title: fix(epay): upgrade go-epay to v0.0.5
- Feature/group: fix(epay): upgrade go-epay to v0.0.5
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/payment_webhook_availability_test.go`, `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bb28f2ccad3f

- Title: feat(tokenkit): move token counting into a module calibrated on agent traffic
- Feature/group: feat(tokenkit): move token counting into a module calibrated on agent traffic
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/workflows/ci.yml`, `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`, `AGENTS.md`, `Dockerfile`, `Dockerfile.dev`, `common/model.go`, `go.mod`, `go.sum`, `main.go`, `makefile`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_claude_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/openai/relay-openai.go`, `relaykit/dto/claude.go`, `relaykit/dto/gemini.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/token_count_meta_test.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_resp.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/token_counter.go`, `service/token_estimator.go`, `service/tokenizer.go`, `service/usage_helpr.go`, `tokenkit/README.md`, `tokenkit/go.mod`, `tokenkit/go.sum`, `tokenkit/image.go`, `tokenkit/image_test.go`, `tokenkit/tokenizer.go`, `tokenkit/tokenkit.go`, `tokenkit/tokenkit_test.go`
- Overlap with downstream changes: .github/workflows/ci.yml, .github/workflows/electron-build.yml, .github/workflows/release.yml, AGENTS.md, Dockerfile, go.mod, go.sum, main.go, makefile, relay/channel/openai/relay-openai.go, relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0a532f7f7d1b

- Title: feat(plugins): add an xAI Grok Imagine video plugin
- Feature/group: feat(plugins): add an xAI Grok Imagine video plugin
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/builtin_plugins_test.go`, `plugins/tasks/xai/plugin.js`, `plugins/xai_responses_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0a03088e0e73

- Title: fix(relayconvert): fix lossy conversions and converted-stream failure billing
- Feature/group: fix(relayconvert): fix lossy conversions and converted-stream failure billing
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relaykit/dto/openai_request.go`, `relaykit/dto/openai_response.go`, `relaykit/relayconvert/convmeta/meta.go`, `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/internal/gemini_chat/to_oai_chat_req.go`, `relaykit/relayconvert/internal/media/media.go`, `relaykit/relayconvert/internal/oai_chat/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_resp.go`, `relaykit/relayconvert/internal/oai_chat/to_oai_responses_stream_resp.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/internal/shared/claude/media.go`, `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/toolconv/decode.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/request_registry.go`
- Overlap with downstream changes: relaykit/dto/openai_request.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-acb347a197ff

- Title: feat(relay): web search encoding per channel and vendor-reported search billing (#7690)
- Feature/group: feat(relay): web search encoding per channel and vendor-reported search billing (#7690)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/alpha_search_handler.go`, `relay/channel/advancedcustom/adaptor.go`, `relay/channel/advancedcustom/adaptor_test.go`, `relay/channel/ali/adaptor.go`, `relay/channel/ali/adaptor_test.go`, `relay/channel/ali/web_search.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/claude/tool_billing_test.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_responses.go`, `relay/channel/moonshot/adaptor.go`, `relay/channel/moonshot/adaptor_test.go`, `relay/channel/openai/adaptor.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/relay_responses_billing_test.go`, `relay/channel/openai/web_search.go`, `relay/channel/xai/adaptor.go`, `relay/channel/xai/web_search.go`, `relay/channel/zhipu_4v/adaptor.go`, `relay/channel/zhipu_4v/web_search.go`, `relay/common/relay_info.go`, `relay/common/tool_usage.go`, `relay/common/tool_usage_test.go`, `relay/response_model_test.go`, `relay/responses_websocket.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/openai_request.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/convmeta/websearch.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/request_registry.go`, `relaykit/relayconvert/request_registry_test.go`, `relaykit/relayconvert/text_converter_registry.go`, `service/responses_usage.go`, `service/responses_usage_test.go`, `service/text_quota.go`, `service/text_quota_test.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `web/src/features/channels/lib/advanced-custom.ts`, `web/src/features/channels/types.ts`, `web/src/features/system-settings/models/tool-price-settings.tsx`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: relay/channel/openai/adaptor.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relaykit/dto/channel_settings.go, relaykit/dto/openai_request.go, service/text_quota.go, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-45094bdf41fb

- Title: fix(web): 修复 OriginOS 6 设备渠道页白屏
- Feature/group: fix(web): 修复 OriginOS 6 设备渠道页白屏
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/lib/motion.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-04c597bf6cce

- Title: feat: support regex conditions in advanced parameter overrides (#7698)
- Feature/group: feat: support regex conditions in advanced parameter overrides (#7698)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/common/override.go`, `relay/common/override_test.go`, `web/src/features/channels/components/__tests__/param-override-editor.test.tsx`, `web/src/features/channels/components/dialogs/param-override-editor-dialog.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-d10f11ffff20

- Title: feat(gemini): bill Google Search grounding per query or per grounded prompt
- Feature/group: feat(gemini): bill Google Search grounding per query or per grounded prompt
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/gemini/relay_gemini_usage_test.go`, `relay/channel/gemini/relay_responses.go`, `relay/common/tool_usage.go`, `relay/common/tool_usage_test.go`, `relaykit/dto/gemini.go`, `service/text_quota_test.go`, `setting/operation_setting/tools.go`, `setting/operation_setting/tools_price_test.go`, `web/src/features/system-settings/models/tool-price-settings.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0cce21db0283

- Title: perf(jsplugin): stop rebuilding the usage-key replacer and request body on every hook call
- Feature/group: perf(jsplugin): stop rebuilding the usage-key replacer and request body on every hook call
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bc823579294f

- Title: test(relaykit): keep format-conversion tests out of the repository
- Feature/group: test(relaykit): keep format-conversion tests out of the repository
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req_test.go`, `relaykit/relayconvert/reasoning/intent_test.go`, `relaykit/relayconvert/reasoning/suffix_test.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b23471112d8f

- Title: build: upgrade moejs to v0.1.0-alpha.6
- Feature/group: build: upgrade moejs to v0.1.0-alpha.6
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-192802ee0691

- Title: build: build release binaries with moejs's PGO profile
- Feature/group: build: build release binaries with moejs's PGO profile
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/workflows/electron-build.yml`, `.github/workflows/release.yml`, `Dockerfile`, `electron/build.sh`
- Overlap with downstream changes: .github/workflows/electron-build.yml, .github/workflows/release.yml, Dockerfile
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-6370b2942416

- Title: feat(jsplugin): keep JSON member order for plugins that declare json-order@1
- Feature/group: feat(jsplugin): keep JSON member order for plugins that declare json-order@1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `middleware/task_plugin.go`, `middleware/task_plugin_test.go`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/routing.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7aa3531ef4c2

- Title: fix(task-plugin): set multipart content type for upstream requests (#7655)
- Feature/group: fix(task-plugin): set multipart content type for upstream requests (#7655)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-e0483122c7d3

- Title: fix(relayconvert): drop trailing assistant text from Responses input on Claude and Gemini
- Feature/group: fix(relayconvert): drop trailing assistant text from Responses input on Claude and Gemini
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/req_helpers.go`, `relaykit/relayconvert/internal/oai_responses/to_claude_messages_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-128089fa7ddf

- Title: feat(i18n): translate web console messages in the frontend and AI-client errors in the backend (#7712)
- Feature/group: feat(i18n): translate web console messages in the frontend and AI-client errors in the backend (#7712)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.env.example`, `AGENTS.md`, `common/audio.go`, `common/body_storage.go`, `common/constants.go`, `common/email.go`, `common/env.go`, `common/gin.go`, `common/gopool.go`, `common/init.go`, `common/limiter/limiter.go`, `common/log_text.go`, `common/log_text.zh-CN.json`, `common/message.go`, `common/pprof.go`, `common/redis.go`, `common/sys_log.go`, `common/topup-ratio.go`, `common/totp.go`, `common/utils.go`, `controller/access_token.go`, `controller/access_token_audit_test.go`, `controller/audit.go`, `controller/auth_session.go`, `controller/channel-billing.go`, `controller/channel-test.go`, `controller/channel.go`, `controller/channel_affinity_cache.go`, `controller/channel_test_internal_test.go`, `controller/channel_upstream_update.go`, `controller/channel_upstream_update_test.go`, `controller/checkin.go`, `controller/codex_usage.go`, `controller/custom_oauth.go`, `controller/group.go`, `controller/log.go`, `controller/login_verification.go`, `controller/midjourney.go`, `controller/misc.go`, `controller/model.go`, `controller/model_meta.go`, `controller/oauth.go`, `controller/option.go`, `controller/passkey.go`, `controller/passkey_test.go`, `controller/payment_compliance.go`, `controller/performance.go`, `controller/playground.go`, `controller/plugin_protocol.go`, `controller/plugin_protocol_image.go`, `controller/prefill_group.go`, `controller/pricing.go`, `controller/ratio_config.go`, `controller/ratio_sync.go`, `controller/redemption.go`, `controller/relay.go`, `controller/responses_websocket.go`, `controller/secure_verification.go`, `controller/security_enrollment_test.go`, `controller/setup.go`, `controller/subscription.go`, `controller/subscription_payment_creem.go`, `controller/subscription_payment_epay.go`, `controller/subscription_payment_stripe.go`, `controller/subscription_payment_waffo_pancake.go`, `controller/system_task_handlers.go`, `controller/task_plugin.go`, `controller/task_plugin_test.go`, `controller/theme_compat_test.go`, `controller/token.go`, `controller/topup.go`, `controller/topup_creem.go`, `controller/topup_quota_limit_test.go`, `controller/topup_stripe.go`, `controller/topup_waffo.go`, `controller/topup_waffo_pancake.go`, `controller/twofa.go`, `controller/usedata.go`, `controller/user.go`, `controller/user_quota.go`, `controller/vendor_meta.go`, `controller/video_proxy.go`, `controller/wechat.go`, `docker-compose.yml`, `i18n/i18n.go`, `i18n/i18n_test.go`, `i18n/keys.go`, `i18n/locales/en.yaml`, `i18n/locales/zh-CN.yaml`, `i18n/locales/zh-TW.yaml`, `logger/logger.go`, `main.go`, `middleware/auth.go`, `middleware/auth_test.go`, `middleware/distributor.go`, `middleware/email-verification-rate-limit.go`, `middleware/i18n.go`, `middleware/model-rate-limit.go`, `middleware/rate-limit.go`, `middleware/rate_limit_test.go`, `middleware/recover.go`, `middleware/secure_verification.go`, `middleware/task_plugin.go`, `middleware/trusted_proxies.go`, `middleware/turnstile-check.go`, `middleware/utils.go`, `model/ability.go`, `model/audit_log.go`, `model/channel.go`, `model/channel_cache.go`, `model/checkin.go`, `model/frontend_option_migration.go`, `model/gorm_logger.go`, `model/log.go`, `model/log_other.go`, `model/main.go`, `model/model_pricing_config.go`, `model/option.go`, `model/option_primary_key_migration.go`, `model/passkey.go`, `model/pricing.go`, `model/quota_reserve.go`, `model/redemption.go`, `model/subscription.go`, `model/subscription_auth_test.go`, `model/subscription_reset_test.go`, `model/task_model_alias.go`, `model/token.go`, `model/topup.go`, `model/twofa.go`, `model/usedata.go`, `model/user.go`, `model/user_access_token.go`, `model/user_cache.go`, `model/user_quota_adjustment.go`, `model/user_session.go`, `model/utils.go`, `oauth/discord.go`, `oauth/generic.go`, `oauth/github.go`, `oauth/linuxdo.go`, `oauth/oidc.go`, `oauth/registry.go`, `oauth/types.go`, `pkg/perf_metrics/flush.go`, `pkg/wsmanager/wsmanager.go`, `relay/audio_handler.go`, `relay/channel/api_request.go`, `relay/channel/aws/relay-aws.go`, `relay/channel/baidu/relay-baidu.go`, `relay/channel/claude/relay-claude.go`, `relay/channel/claude/relay_responses.go`, `relay/channel/cloudflare/relay_cloudflare.go`, `relay/channel/cohere/relay-cohere.go`, `relay/channel/coze/relay-coze.go`, `relay/channel/dify/relay-dify.go`, `relay/channel/gemini/relay-gemini-native.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/jimeng/sign.go`, `relay/channel/ollama/relay-ollama.go`, `relay/channel/ollama/stream.go`, `relay/channel/openai/audio.go`, `relay/channel/openai/chat_via_responses.go`, `relay/channel/openai/helper.go`, `relay/channel/openai/relay-openai.go`, `relay/channel/openai/relay_image.go`, `relay/channel/openai/relay_realtime.go`, `relay/channel/openai/relay_responses.go`, `relay/channel/openai/responses_via_chat.go`, `relay/channel/palm/relay-palm.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/tencent/relay-tencent.go`, `relay/channel/xai/text.go`, `relay/channel/xunfei/relay-xunfei.go`, `relay/channel/zhipu/relay-zhipu.go`, `relay/channel/zhipu_4v/image.go`, `relay/common/conversion_diagnostics.go`, `relay/common/relay_info.go`, `relay/common/tool_usage.go`, `relay/compatible_handler.go`, `relay/gemini_handler.go`, `relay/helper/common.go`, `relay/helper/price.go`, `relay/helper/stream_scanner.go`, `relay/helper/valid_request.go`, `relay/image_handler.go`, `relay/mjproxy_handler.go`, `relay/relay_task.go`, `relay/relay_task_test.go`, `relay/request_billing.go`, `relay/responses_handler.go`, `relay/websocket.go`, `router/main.go`, `router/plugin-router.go`, `service/auth_cleanup.go`, `service/authz/adapter.go`, `service/authz/enforcer.go`, `service/billing.go`, `service/billing_session.go`, `service/channel.go`, `service/channel_affinity.go`, `service/channel_select.go`, `service/codex_credential_refresh_task.go`, `service/download.go`, `service/error.go`, `service/file_decoder.go`, `service/file_service.go`, `service/group.go`, `service/http.go`, `service/http_client.go`, `service/http_transport_policy.go`, `service/http_transport_sharded.go`, `service/image.go`, `service/log_info_generate.go`, `service/midjourney.go`, `service/passkey/service.go`, `service/passkey/session.go`, `service/quota.go`, `service/relay_error.go`, `service/subscription_reset_task.go`, `service/system_instance.go`, `service/system_task.go`, `service/task_billing.go`, `service/task_billing_test.go`, `service/task_polling.go`, `service/task_polling_test.go`, `service/text_quota.go`, `service/text_quota_test.go`, `service/tiered_settle.go`, `service/tiered_settle_test.go`, `service/user_notify.go`, `service/violation_fee.go`, `service/webhook.go`, `setting/chat.go`, `setting/config/config.go`, `setting/console_setting/validation.go`, `setting/model_setting/global.go`, `setting/operation_setting/tools.go`, `setting/rate_limit.go`, `setting/ratio_setting/group_ratio.go`, `setting/ratio_setting/model_ratio.go`, `setting/system_setting/task_artifact.go`, `setting/system_setting/task_artifact_store.go`, `setting/user_usable_group.go`, `web/src/components/model-group-selector.tsx`, `web/src/components/model-group-selector/__tests__/group-detail.test.ts`, `web/src/components/model-group-selector/group-detail.ts`, `web/src/features/auth/api.ts`, `web/src/features/auth/passkey/hooks/use-passkey-management.ts`, `web/src/features/channels/components/channels-columns.tsx`, `web/src/features/channels/components/dialogs/codex-usage-dialog.tsx`, `web/src/features/channels/components/dialogs/multi-key-manage-dialog.tsx`, `web/src/features/channels/components/dialogs/ollama-models-dialog.tsx`, `web/src/features/channels/lib/channel-actions.ts`, `web/src/features/keys/components/api-keys-mutate-drawer.tsx`, `web/src/features/models/components/dialogs/model-delete-dialog.tsx`, `web/src/features/models/components/models-table.tsx`, `web/src/features/playground/api.ts`, `web/src/features/playground/types.ts`, `web/src/features/subscriptions/components/dialogs/user-subscriptions-dialog.tsx`, `web/src/features/subscriptions/components/subscriptions-mutate-drawer.tsx`, `web/src/features/system-settings/auth/passkey-section.tsx`, `web/src/features/system-settings/integrations/payment-settings-section.tsx`, `web/src/features/system-settings/integrations/waffo-pancake-settings-section.tsx`, `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/column-helpers.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`, `web/src/features/usage-logs/components/columns/task-logs-columns.tsx`, `web/src/features/usage-logs/components/common-log-mobile-card.tsx`, `web/src/features/usage-logs/components/dialogs/details-dialog.tsx`, `web/src/features/usage-logs/components/dialogs/task-details-dialog.tsx`, `web/src/features/usage-logs/lib/__tests__/token-name.test.ts`, `web/src/features/usage-logs/lib/format.ts`, `web/src/features/usage-logs/types.ts`, `web/src/features/wallet/hooks/use-affiliate.ts`, `web/src/features/wallet/hooks/use-waffo-pancake-payment.ts`, `web/src/features/wallet/hooks/use-waffo-payment.ts`, `web/src/features/wallet/lib/payment.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`, `web/src/lib/secure-verification.ts`, `web/src/lib/server-error-message.test.ts`, `web/src/lib/server-error-message.ts`
- Overlap with downstream changes: .env.example, AGENTS.md, common/constants.go, common/gin.go, common/redis.go, controller/audit.go, controller/channel-test.go, controller/channel.go, controller/log.go, controller/misc.go, controller/model.go, controller/option.go, controller/playground.go, controller/relay.go, controller/secure_verification.go, controller/system_task_handlers.go, controller/user.go, controller/video_proxy.go, docker-compose.yml, i18n/keys.go, i18n/locales/en.yaml, i18n/locales/zh-CN.yaml, i18n/locales/zh-TW.yaml, logger/logger.go, main.go, middleware/distributor.go, model/ability.go, model/channel.go, model/channel_cache.go, model/gorm_logger.go, model/log.go, model/main.go, model/option.go, model/twofa.go, model/usedata.go, model/user.go, model/user_cache.go, relay/channel/openai/helper.go, relay/channel/openai/relay-openai.go, relay/common/relay_info.go, relay/compatible_handler.go, relay/helper/common.go, relay/mjproxy_handler.go, relay/relay_task.go, relay/responses_handler.go, service/auth_cleanup.go, service/billing.go, service/billing_session.go, service/download.go, service/http.go, service/http_client.go, service/midjourney.go, service/quota.go, service/system_task.go, service/task_billing.go, service/task_billing_test.go, service/task_polling.go, service/task_polling_test.go, service/text_quota.go, service/violation_fee.go, service/webhook.go, setting/console_setting/validation.go, setting/ratio_setting/model_ratio.go, web/src/features/usage-logs/components/columns/common-logs-columns.tsx, web/src/features/usage-logs/components/columns/task-logs-columns.tsx, web/src/features/usage-logs/components/dialogs/details-dialog.tsx, web/src/features/usage-logs/lib/format.ts, web/src/features/usage-logs/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json, web/src/lib/server-error-message.test.ts, web/src/lib/server-error-message.ts
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1d4328e97417

- Title: docs(relaykit): rewrite README for the current conversion API
- Feature/group: docs(relaykit): rewrite README for the current conversion API
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/README.md`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b66674ac23d7

- Title: feat(task-plugin): let Seedance and wan3.0 pick the video duration with -1
- Feature/group: feat(task-plugin): let Seedance and wan3.0 pick the video duration with -1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.agents/rules/billing.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `docs/plugin-api/v1.schema.json`, `pkg/jsplugin/json_state.go`, `pkg/jsplugin/registry.go`, `plugins/alibaba_wan3_test.go`, `plugins/doubao_responses_test.go`, `plugins/tasks/alibaba/plugin.js`, `plugins/tasks/doubao/plugin.js`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-29509776ef85

- Title: feat(jsplugin): let driver hooks read upstream data with utils.fetch
- Feature/group: feat(jsplugin): let driver hooks read upstream data with utils.fetch
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `common/log_text.zh-CN.json`, `controller/task.go`, `controller/video_proxy.go`, `docs/plugin-api/README.md`, `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `pkg/jsplugin/engine.go`, `pkg/jsplugin/engine_test.go`, `pkg/jsplugin/fetch.go`, `pkg/jsplugin/request.go`, `pkg/jsplugin/utils.go`, `plugins/alibaba_responses_test.go`, `plugins/hailuo_responses_test.go`, `plugins/xai_responses_test.go`, `relay/channel/adapter.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/auth_test.go`
- Overlap with downstream changes: controller/task.go, controller/video_proxy.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-66c3e142820e

- Title: fix(hailuo): download MiniMax v1 videos through the file retrieve link
- Feature/group: fix(hailuo): download MiniMax v1 videos through the file retrieve link
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/task_generic_test.go`, `plugins/hailuo_responses_test.go`, `plugins/tasks/hailuo/plugin.js`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3f02244783e0

- Title: feat(web): show the request domains a plugin declares when uploading it
- Feature/group: feat(web): show the request domains a plugin declares when uploading it
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/task-plugins/__tests__/upload-dialog.test.tsx`, `web/src/features/task-plugins/components/marketplace-capabilities.tsx`, `web/src/features/task-plugins/components/upload-dialog.tsx`, `web/src/features/task-plugins/lib/plugin-meta-preview.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ba3b2c015ad8

- Title: refactor(jsplugin): stop requiring json-clone@1 and submit-sse-delta@1
- Feature/group: refactor(jsplugin): stop requiring json-clone@1 and submit-sse-delta@1
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `docs/plugin-api/v1.d.ts`, `docs/plugin-api/v1.md`, `pkg/jsplugin/registry.go`, `pkg/jsplugin/registry_test.go`, `plugins/tasks/alibaba/plugin.js`, `relay/channel/task/jsplugin/adaptor_test.go`, `relay/channel/task/jsplugin/submit_stream.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c8d234baa18b

- Title: fix(xai): bill video edits and extensions by their output resolution
- Feature/group: fix(xai): bill video edits and extensions by their output resolution
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `plugins/tasks/xai/plugin.js`, `plugins/xai_responses_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c6741c36a9a1

- Title: feat(channel): add a per-channel switch to skip upstream TLS verification
- Feature/group: feat(channel): add a per-channel switch to skip upstream TLS verification
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/channel-billing.go`, `controller/channel_upstream_update.go`, `controller/codex_usage.go`, `controller/video_proxy.go`, `relay/channel/api_request.go`, `relay/channel/gemini/relay-gemini.go`, `relay/channel/task/jsplugin/adaptor.go`, `relay/channel/vertex/service_account.go`, `relay/mjproxy_handler.go`, `relaykit/dto/channel_settings.go`, `relaykit/dto/channel_settings_test.go`, `service/codex_channel_models.go`, `service/http_client.go`, `service/http_client_transport_test.go`, `service/http_transport_policy.go`, `service/task_polling.go`, `web/src/features/channels/components/__tests__/channel-configuration.test.tsx`, `web/src/features/channels/components/drawers/channel-mutate-drawer.tsx`, `web/src/features/channels/lib/channel-configuration.ts`, `web/src/features/channels/lib/channel-form-errors.ts`, `web/src/features/channels/lib/channel-form.ts`, `web/src/features/channels/types.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/video_proxy.go, relay/mjproxy_handler.go, relaykit/dto/channel_settings.go, relaykit/dto/channel_settings_test.go, service/http_client.go, service/http_client_transport_test.go, service/task_polling.go, web/src/features/channels/components/drawers/channel-mutate-drawer.tsx, web/src/features/channels/lib/channel-form-errors.ts, web/src/features/channels/lib/channel-form.ts, web/src/features/channels/types.ts, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1ca79d57e77d

- Title: fix(pricing): show recent status samples without gaps
- Feature/group: fix(pricing): show recent status samples without gaps
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/perf_metrics.go`, `model/perf_metric.go`, `pkg/perf_metrics/metrics.go`, `pkg/perf_metrics/metrics_test.go`, `pkg/perf_metrics/types.go`, `web/src/features/performance-metrics/api.ts`, `web/src/features/performance-metrics/types.ts`, `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card-grid.tsx`, `web/src/features/pricing/components/model-perf-badge.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-0f38a14a5e8e

- Title: fix(web): restore browser sans-serif for Auto font
- Feature/group: fix(web): restore browser sans-serif for Auto font
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/config-drawer.tsx`, `web/src/context/__tests__/theme-preferences.test.tsx`, `web/src/context/theme-customization-provider.tsx`, `web/src/lib/theme-customization.ts`, `web/src/styles/index.css`, `web/src/styles/theme-presets.css`, `web/src/styles/theme.css`
- Overlap with downstream changes: web/src/styles/index.css, web/src/styles/theme.css
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-036ba376da7d

- Title: fix(relaykit): match Claude effort and Gemini 2.5 budgets to each model
- Feature/group: fix(relaykit): match Claude effort and Gemini 2.5 budgets to each model
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/shared/claude/reasoning.go`, `relaykit/relayconvert/reasoning/claude.go`, `relaykit/relayconvert/reasoning/gemini.go`, `relaykit/relayconvert/reasoning/suffix.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-13c8e90b8f4b

- Title: fix(relayconvert): send Responses image URLs to Chat upstreams as objects
- Feature/group: fix(relayconvert): send Responses image URLs to Chat upstreams as objects
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_responses/to_oai_chat_req.go`, `relaykit/relayconvert/testdata/golden/request/openai_responses_to_openai.golden.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-5ac75628409d

- Title: fix(relayconvert): keep answer text that shares a stream chunk with reasoning
- Feature/group: fix(relayconvert): keep answer text that shares a stream chunk with reasoning
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-977450a3e810

- Title: fix(rate-limit): explain model rate-limit rejections without Redis
- Feature/group: fix(rate-limit): explain model rate-limit rejections without Redis
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket.go`, `controller/responses_websocket_test.go`, `middleware/model-rate-limit.go`, `middleware/model_rate_limit_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b892a6b2e44d

- Title: fix(web): reload once when a page's code fails to load
- Feature/group: fix(web): reload once when a page's code fails to load
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/errors/__tests__/general-error.test.tsx`, `web/src/features/errors/general-error.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1fb6d07fee01

- Title: perf(web): stop loading the whole Simple Icons pack on four pages
- Feature/group: perf(web): stop loading the whole Simple Icons pack on four pages
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/assets/brand-icons/icon-alipay.tsx`, `web/src/assets/brand-icons/index.ts`, `web/src/features/auth/components/oauth-callback-screen.tsx`, `web/src/features/security/components/account-bindings.tsx`, `web/src/features/users/components/dialogs/__tests__/user-binding-dialog.test.tsx`, `web/src/features/users/components/dialogs/user-binding-dialog.tsx`, `web/src/features/wallet/lib/__tests__/payment-icons.test.tsx`, `web/src/features/wallet/lib/ui.tsx`
- Overlap with downstream changes: web/src/features/users/components/dialogs/user-binding-dialog.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-06dd76a9d38e

- Title: fix(playground): keep each account's Playground data separate
- Feature/group: fix(playground): keep each account's Playground data separate
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/playground/lib/storage/__tests__/storage.test.ts`, `web/src/features/playground/lib/storage/storage.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-96275d224bfc

- Title: fix(web): show date pickers and relative times in the interface language
- Feature/group: fix(web): show date pickers and relative times in the interface language
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/__tests__/date-picker-locale.test.tsx`, `web/src/components/date-picker.tsx`, `web/src/components/datetime-picker.tsx`, `web/src/features/security/components/__tests__/login-session-item.test.tsx`, `web/src/features/security/components/login-session-item.tsx`, `web/src/features/security/components/passkey-card.tsx`, `web/src/features/system-settings/content/announcements-section.tsx`
- Overlap with downstream changes: web/src/features/system-settings/content/announcements-section.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-93ea55f3c310

- Title: feat(web): pick dates and times down to the second
- Feature/group: feat(web): pick dates and times down to the second
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/components/__tests__/datetime-picker-time.test.tsx`, `web/src/components/datetime-picker.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/compact-date-time-range-picker.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/compact-date-time-range-picker.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4b13d9388ed6

- Title: fix(relayconvert): drop non-string enum values when converting OpenAI tool schemas (#7729)
- Feature/group: fix(relayconvert): drop non-string enum values when converting OpenAI tool schemas (#7729)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/shared/gemini/schema_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-801456a061fb

- Title: fix(relay): send a close frame when a Responses WebSocket session ends
- Feature/group: fix(relay): send a close frame when a Responses WebSocket session ends
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/responses_websocket_test.go`, `relay/responses_websocket.go`, `relay/responses_websocket_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c10780e054e6

- Title: fix(relayconvert): keep JSON Schema constraints in tools and structured output sent to Gemini
- Feature/group: fix(relayconvert): keep JSON Schema constraints in tools and structured output sent to Gemini
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/common/relay_info.go`, `relay/common/relay_info_gemini_settings_test.go`, `relaykit/relayconvert/convmeta/options.go`, `relaykit/relayconvert/internal/oai_chat/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/oai_responses/to_gemini_chat_req.go`, `relaykit/relayconvert/internal/shared/gemini/schema.go`, `relaykit/relayconvert/internal/shared/gemini/schema_test.go`, `relaykit/relayconvert/internal/toolconv/encode.go`, `relaykit/relayconvert/internal/toolconv/model.go`, `relaykit/relayconvert/request_registry.go`
- Overlap with downstream changes: relay/common/relay_info.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-42e694eb60fe

- Title: fix(web): map Meta Muse Spark models to Meta icon (#7734)
- Feature/group: fix(web): map Meta Muse Spark models to Meta icon (#7734)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/model-badge.test.tsx`, `web/src/lib/model-provider.ts`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3827fd83076b

- Title: fix(web): show cache read price on expression model cards (#7757)
- Feature/group: fix(web): show cache read price on expression model cards (#7757)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/pricing/__tests__/model-cards.test.tsx`, `web/src/features/pricing/components/model-card.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3f926c3ef497

- Title: fix(metrics): count completed streams after context cancellation (#7638)
- Feature/group: fix(metrics): count completed streams after context cancellation (#7638)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `pkg/perf_metrics/outcome.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bbf60b49f6a4

- Title: fix(logs): vertically align billing source icons on mobile (#7760)
- Feature/group: fix(logs): vertically align billing source icons on mobile (#7760)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/cost-display.test.tsx`, `web/src/features/usage-logs/components/log-cost-display.tsx`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b22a43b6e89e

- Title: fix: mark task settlement logs with is_task for tiered billing display (#7446)
- Feature/group: fix: mark task settlement logs with is_task for tiered billing display (#7446)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `service/task_billing.go`
- Overlap with downstream changes: service/task_billing.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-34b19421237f

- Title: fix(pricing): allow root access when the model square is disabled (#7761)
- Feature/group: fix(pricing): allow root access when the model square is disabled (#7761)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `middleware/header_nav.go`, `middleware/header_nav_test.go`
- Overlap with downstream changes: middleware/header_nav.go
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ba0e478e061e

- Title: fix(relaykit): preserve assistant text with tool calls (#6871)
- Feature/group: fix(relaykit): preserve assistant text with tool calls (#6871)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relaykit/relayconvert/internal/claude_messages/to_oai_chat_req.go`, `relaykit/relayconvert/request_registry_test.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-fc08b76f0fb3

- Title: fix(web): preserve recorded billing tiers in usage log previews (#7763)
- Feature/group: fix(web): preserve recorded billing tiers in usage log previews (#7763)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/usage-logs/components/__tests__/detail-preview.test.tsx`, `web/src/features/usage-logs/components/columns/common-logs-columns.tsx`
- Overlap with downstream changes: web/src/features/usage-logs/components/columns/common-logs-columns.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-4bfcf2812198

- Title: chore(web): bump @lobehub/icons to 5.22.0 (#7764)
- Feature/group: chore(web): bump @lobehub/icons to 5.22.0 (#7764)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9948b1ef14ea

- Title: fix(pricing): accept declared endpoint arrays in pricing conversion (#7724)
- Feature/group: fix(pricing): accept declared endpoint arrays in pricing conversion (#7724)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/model_pricing_conversion_endpoints_test.go`, `model/model_pricing_conversion.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2ed058795bf0

- Title: fix(ci): stabilize checks and configure Bun dependency updates (#7765)
- Feature/group: fix(ci): stabilize checks and configure Bun dependency updates (#7765)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `.github/dependabot.yml`, `.github/workflows/ci.yml`, `controller/security_account_test.go`, `controller/security_enrollment_test.go`, `web/bun.lock`, `web/package.json`, `web/src/components/ui/__tests__/markdown.test.tsx`, `web/src/features/channels/components/__tests__/model-mapping-editor.test.tsx`
- Overlap with downstream changes: .github/workflows/ci.yml
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-2b85a8fc67bd

- Title: fix(auth): show legal consent warning on password login (#7771)
- Feature/group: fix(auth): show legal consent warning on password login (#7771)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/src/features/auth/sign-in/components/user-auth-form.tsx`
- Overlap with downstream changes: web/src/features/auth/sign-in/components/user-auth-form.tsx
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-767a82b87647

- Title: fix(claude): forward thinking.block_binding and every anthropic-beta value (#7732)
- Feature/group: fix(claude): forward thinking.block_binding and every anthropic-beta value (#7732)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `relay/channel/aws/dto.go`, `relay/channel/aws/relay_aws_test.go`, `relay/channel/claude/adaptor.go`, `relay/channel/claude/adaptor_test.go`, `relay/channel/claude/relay_claude_test.go`, `relaykit/dto/claude.go`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-ba29f0645394

- Title: chore(deps-dev): bump @xmldom/xmldom from 0.8.13 to 0.8.15 in /electron (#7187)
- Feature/group: chore(deps-dev): bump @xmldom/xmldom from 0.8.13 to 0.8.15 in /electron (#7187)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-b8c2dc12660d

- Title: build(deps): bump yace from 1.1.0 to 1.1.1 in /web (#7770)
- Feature/group: build(deps): bump yace from 1.1.0 to 1.1.1 in /web (#7770)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-c730add954e1

- Title: build(deps): bump stream-markdown-parser from 1.1.3 to 1.2.18 in /web (#7769)
- Feature/group: build(deps): bump stream-markdown-parser from 1.1.3 to 1.2.18 in /web (#7769)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-82306b3c3594

- Title: build(deps-dev): bump @tanstack/router-plugin in /web (#7768)
- Feature/group: build(deps-dev): bump @tanstack/router-plugin in /web (#7768)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-bf4e1ebfb885

- Title: build(deps-dev): bump oxfmt from 0.57.0 to 0.72.0 in /web (#7766)
- Feature/group: build(deps-dev): bump oxfmt from 0.57.0 to 0.72.0 in /web (#7766)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7e25797d1468

- Title: chore(deps-dev): bump electron from 39.8.10 to 41.10.6 in /electron (#7625)
- Feature/group: chore(deps-dev): bump electron from 39.8.10 to 41.10.6 in /electron (#7625)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`, `electron/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-7f51a97226c4

- Title: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Feature/group: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-1c58f3d161fa

- Title: chore(deps-dev): bump fast-uri from 3.1.5 to 3.1.7 in /electron (#7169)
- Feature/group: chore(deps-dev): bump fast-uri from 3.1.5 to 3.1.7 in /electron (#7169)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3fe394cdba28

- Title: chore(deps): bump github.com/gorilla/websocket from 1.5.0 to 1.5.3 (#7001)
- Feature/group: chore(deps): bump github.com/gorilla/websocket from 1.5.0 to 1.5.3 (#7001)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `go.mod`, `go.sum`
- Overlap with downstream changes: go.mod, go.sum
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-9f048107edbb

- Title: chore(deps): bump ai from 7.0.31 to 7.0.136 in /web (#7767)
- Feature/group: chore(deps): bump ai from 7.0.31 to 7.0.136 in /web (#7767)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `web/bun.lock`, `web/package.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-3c9fb2a0de16

- Title: chore(deps-dev): bump brace-expansion from 1.1.16 to 1.1.21 in /electron (#7623)
- Feature/group: chore(deps-dev): bump brace-expansion from 1.1.16 to 1.1.21 in /electron (#7623)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-93cf4ba6fb92

- Title: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Feature/group: chore(deps-dev): bump js-yaml from 4.3.1 to 4.3.2 in /electron (#7327)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: (none)
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-dd59854050f2

- Title: chore(deps-dev): bump undici from 6.28.0 to 6.29.0 in /electron (#7624)
- Feature/group: chore(deps-dev): bump undici from 6.28.0 to 6.29.0 in /electron (#7624)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `electron/package-lock.json`
- Overlap with downstream changes: (none detected)
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.

## commit-633b4c5b190d

- Title: feat(logs): export usage logs from the common log page (#7774)
- Feature/group: feat(logs): export usage logs from the common log page (#7774)
- Decision: `pending`
- Selected scope: `full_commit`
- Reason: 
- Dependencies: none declared; requires review
- Files: `controller/log.go`, `controller/log_export_test.go`, `controller/misc.go`, `controller/relay_error_log_test.go`, `controller/usedata.go`, `controller/user_manage_test.go`, `middleware/access_token_routes.go`, `model/log.go`, `model/option.go`, `router/api-router.go`, `service/log_export.go`, `setting/operation_setting/log_setting.go`, `web/src/features/auth/types.ts`, `web/src/features/system-settings/hooks/use-update-option.ts`, `web/src/features/system-settings/maintenance/__tests__/log-settings.test.tsx`, `web/src/features/system-settings/maintenance/history-log-cleanup.tsx`, `web/src/features/system-settings/maintenance/log-settings-section.tsx`, `web/src/features/system-settings/maintenance/server-log-files.tsx`, `web/src/features/system-settings/operations/index.tsx`, `web/src/features/system-settings/operations/section-registry.tsx`, `web/src/features/system-settings/types.ts`, `web/src/features/usage-logs/api.ts`, `web/src/features/usage-logs/components/__tests__/export-dialog.test.tsx`, `web/src/features/usage-logs/components/__tests__/mobile-filter.test.tsx`, `web/src/features/usage-logs/components/common-logs-filter-bar.tsx`, `web/src/features/usage-logs/components/dialogs/export-logs-dialog.tsx`, `web/src/features/usage-logs/components/logs-filter-toolbar.tsx`, `web/src/features/usage-logs/lib/__tests__/export.test.ts`, `web/src/features/usage-logs/lib/export.ts`, `web/src/i18n/locales/en.json`, `web/src/i18n/locales/fr.json`, `web/src/i18n/locales/ja.json`, `web/src/i18n/locales/ru.json`, `web/src/i18n/locales/vi.json`, `web/src/i18n/locales/zh-TW.json`, `web/src/i18n/locales/zh.json`
- Overlap with downstream changes: controller/log.go, controller/misc.go, controller/user_manage_test.go, model/log.go, model/option.go, router/api-router.go, web/src/features/auth/types.ts, web/src/features/system-settings/hooks/use-update-option.ts, web/src/features/system-settings/maintenance/log-settings-section.tsx, web/src/features/system-settings/operations/section-registry.tsx, web/src/features/system-settings/types.ts, web/src/features/usage-logs/api.ts, web/src/features/usage-logs/components/common-logs-filter-bar.tsx, web/src/i18n/locales/en.json, web/src/i18n/locales/fr.json, web/src/i18n/locales/ja.json, web/src/i18n/locales/ru.json, web/src/i18n/locales/vi.json, web/src/i18n/locales/zh-TW.json, web/src/i18n/locales/zh.json
- Equivalent downstream patch IDs: (none detected)
- Existing source reference: False
- Review behavior, protocol/billing/auth boundaries, all database dialects, dependencies and regression coverage.
