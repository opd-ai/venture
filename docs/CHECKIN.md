Generate a professional git commit message based on the current git diff, then commit and push the changes.

EXECUTION MODE: Autonomous action

STEPS:
1. Run `git diff` to analyze staged/unstaged changes
2. Generate a commit message following conventional commit format:
   - Use type prefixes: feat, fix, docs, refactor, test, perf, chore
   - Format: `<type>(<scope>): <description>`
   - Keep subject line under 72 characters
   - Include bullet points for multiple changes if needed
3. Execute: `git commit -am "<generated_message>"`
4. Execute: `git push --all`

SUCCESS CRITERIA:
- Commit message accurately describes changes
- Commit succeeds without errors
- Push completes successfully

OUTPUT FORMAT:
- Show the generated commit message
- Confirm commit SHA
- Confirm push status