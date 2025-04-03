# Creating a Release

> [!NOTE]  
> The Github API will rate limit cliff quickly. Use a Github PAT and set it as the environment variable `GITHUB_TOKEN` before running this script.


1. Run the [release script][1]: `./scripts/release.sh v[X.Y.Z]`
2. Push the changes: `git push`
3. Check if [Continuous Integration][2] workflow is completed successfully.
4. Push the tags: `git push --tags`
5. Wait for [Continuous Deployment][2] workflow to finish.

[1]: ../scripts/release.sh
[2]: https://github.com/margined-protocol/locust-core/actions
