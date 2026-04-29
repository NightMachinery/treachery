# Repository instructions

- Always commit (and push) changes when you reach a natural endpoint UNLESS there is no git repo already. You can (and should) commit your changes in atomic groups when implementing multiple different feature/fixes at the same time.
- Keep `./docs/` updated as you investigate and change things.
- Do not run ChromeHeadless/Karma browser tests in this environment unless the user explicitly asks for them. Prefer build, lint, format checks, and non-browser tests for routine validation.
