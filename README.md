## Features

<img width="480" height="635" alt="Screenshot_20260820_105535" src="https://github.com/user-attachments/assets/6e2cc20d-b2a2-4e6e-b1cc-7a21838ef844" />


- Support for each event type as documented for [this](https://docs.github.com/en/rest/using-the-rest-api/github-event-types?apiVersion=2026-03-10) version of the GitHub REST API
- Compact rendering of closely related events
- Unique color coding for each type of action
- Comments quoted in full

## Install

Have Go 1.25+ installed, clone the repo and build:
```bash
git clone https://github.com/pkurchanov/ghpeek
cd ghpeek
go build .
```

A system-appropriate `ghpeek` executable should spawn in your current working directory.

## Usage

```bash
./ghpeek <username>

```
If there's such a user on GitHub, magic should happen.

## Rationale

One day I was sick and tired of not being particularly good at any one thing, programming-wise, so I settled on backend development. [This roadmap](https://roadmap.sh/backend) seemed like a nice enough scaffolding to hang my practice on. [This project idea](https://roadmap.sh/projects/github-user-activity) proved a formidable place to actually start.
