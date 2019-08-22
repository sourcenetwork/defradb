workflow "New workflow" {
  on = "push"
  resolves = ["TODO Issue Gen"]
}

action "TODO Issue Gen" {
  uses = "jasonetco/todo@master"
  secrets = ["GITHUB_TOKEN"]
}
