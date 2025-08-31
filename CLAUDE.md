- always use alias @/* imports wherever possible, avoid relative imports unless necessary
- stop creating files or routes just for demo. just tell user how it works
- always use tailwind css, not inline css
- always check your work against linting / formatting rules in biome.jsonc in this project
- always use bun or bunx

# Troubleshooting
- **esbuild EPIPE errors**: `bun remove esbuild && bun add -D esbuild` (native binary corruption fix)

# Git Worktree Management
Whenever you are planning a really big feature or a very complicated fix (everything that is not a one-off or a hotfix), propose to the user whether to create a new Git worktree under the worktrees/ folder and name the feature and worktree with git conventions, e.g., feature/bla-bla-bla or fix/bla-bla-bla. If slash is not possible for worktree name, then change it with a dash separator in the folder name. Also, have a .gitignore to inherit the root directory .gitignore. Never do this yourself. Do not proactively do this, but proactively offer this. And only with user approval will you do this, so users are aware of all of the open work trees and branches.