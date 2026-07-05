## Rationale
One day I was sick and tired of not being particularly good at any one thing, programming-wise, so I settled on backend development. [This roadmap](https://roadmap.sh/backend) seemed like a nice enough scaffolding to hang my practice on. [This project idea](https://roadmap.sh/projects/github-user-activity) seemed a formidable place to actually start.
## Install
Have Go 1.21+ installed, clone the repo and build:
```bash
git clone https://github.com/pkurchanov/ghpeek
cd ghpeek
go build .
```
A system-appropriate `ghpeek` executable should spawn right beside you.
## Usage
```bash
./ghpeek <username>
```
If there's such a user on GitHub, magic should happen.
