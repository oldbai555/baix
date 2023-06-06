git reset "$(git merge-base main "$(git branch --show-current)")"
git add -A && git commit -m 'v0.0.2'
#git push --force