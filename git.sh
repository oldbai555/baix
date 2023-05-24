git reset "$(git merge-base main "$(git branch --show-current)")"
git add -A && git commit -m 'ok'
#git push --force