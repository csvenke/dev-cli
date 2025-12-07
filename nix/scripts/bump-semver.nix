{
  writeShellApplication,
  git,
  semantic-release,
}:

writeShellApplication {
  name = "bump-semver";
  runtimeInputs = [
    git
    semantic-release
  ];
  text = ''
    semantic-release \
    	--branch main \
    	--plugins @semantic-release/commit-analyzer
  '';
}
