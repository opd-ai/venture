Generate a professional git commit message based on the current git diff, then commit and push the changes.

EXECUTION MODE: Autonomous action

STEPS:
1. Run `git diff` and `git status` to analyze staged/unstaged changes
2. Ensure that planning documents such as `AUDIT.md` `PLAN.md` and `ROADMAP.md` are up-to-date without making unnecessary changes.
3. Generate a commit message following conventional commit format:
   - Use type prefixes: feat, fix, docs, refactor, test, perf, chore
   - Format: `<type>(<scope>): <description>`
   - Keep subject line under 72 characters
   - Include bullet points for multiple changes if needed
4. Execute: `git commit -am "<generated_message>"`
5. Execute: `git push --all`

SUCCESS CRITERIA:
- Commit message accurately describes changes
- Commit succeeds without errors
- Push completes successfully

OUTPUT FORMAT:
- Show the generated commit message
- Confirm commit SHA
- Confirm push status
