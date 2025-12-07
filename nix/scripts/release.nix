{
  writeShellApplication,
  git,
  svu,
  goreleaser,
}:

writeShellApplication {
  name = "release";
  runtimeInputs = [
    git
    svu
    goreleaser
  ];
  text = ''
    NEXT_TAG=$(svu next)

    if [ -n "$(git tag -l "$NEXT_TAG")" ]; then
      echo "Tag $NEXT_TAG already exists, skipping release"
      exit 0
    fi

    git tag "$NEXT_TAG"
    git push --tags
    goreleaser release --clean
  '';
}
