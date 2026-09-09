# Billing document font

The statement renderer embeds a static regular instance of Noto Sans SC.
Source: https://github.com/google/fonts/tree/main/ofl/notosanssc
License: SIL Open Font License 1.1 (see OFL.txt).

The font is embedded/subsetted into each PDF. No host font installation, browser,
Python runtime, or network font download is required in production.

`platform-mark-v2.png` is an unchanged copy of `web/public/logo.png`, frozen for
the version 2 statement template. It is combined with the platform name from the
frozen snapshot; rendering never fetches mutable remote assets. Keep old template
assets unchanged so archive retries and confirmation receipts remain deterministic.
