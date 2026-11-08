## Features

<img width="480" height="635" alt="Screenshot_20260820_105535" src="https://github.com/user-attachments/assets/6e2cc20d-b2a2-4e6e-b1cc-7a21838ef844" />


- Support for each event type as documented for [this](https://docs.github.com/en/rest/using-the-rest-api/github-event-types?apiVersion=2026-03-10) version of the GitHub REST API
- Compact rendering of same-type events
- Unique color coding for each type of action
- PR/issue comments quoted in full

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

If there's such a user on GitHub, an overview of their latest activity should appear.
