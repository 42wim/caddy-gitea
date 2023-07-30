# caddy-gitea

[Gitea](https://gitea.io) plugin for [Caddy v2](https://github.com/caddyserver/caddy).

This allows you to have github pages (with more features) in Gitea.
This also requires you to setup a wildcard CNAME to your gitea host.

For now markdown files (with `.md` extension) will also be automatically generated to HTML.

<!-- TOC -->

- [caddy-gitea](#caddy-gitea)
    - [Getting started](#getting-started)
        - [Caddy config](#caddy-config)
        - [DNS config](#dns-config)
        - [Gitea config](#gitea-config)
            - [gitea-pages repo](#gitea-pages-repo)
            - [any repo with configurable allowed branch/tag/commits](#any-repo-with-configurable-allowed-branchtagcommits)
            - [any repo with all branches/tags/commits exposed](#any-repo-with-all-branchestagscommits-exposed)
    - [Building caddy](#building-caddy)

<!-- /TOC -->

## Getting started

### Caddy config

The Caddyfile below creates a webserver listening on :3000 which will interact with gitea on <https://yourgitea.yourdomain.com> using `agiteatoken` as the token.
The agiteatoken should be a token from gitea that has the necessary read rights on the repo's that you want to expose.

```Caddyfile
{
        order gitea before file_server
}
:3000
gitea {
        server https://yourgitea.yourdomain.com
        token agiteatoken
        domain pages.yourdomain.com #this is optional
}
```

### DNS config

This works with a wildcard domain. So you'll need to make a *.pages.yourdomain.com CNAME to the server you'll be running caddy on.
(this doesn't need to be the same server as gitea).

Depending on the gitea config below you'll be able to access your pages using:

- <http://org.pages.yourdomain.com:3000/repo/file.html> (org is the organization or username)
- <http://org.pages.yourdomain.com:3000/repo/file.html?ref=abranch> (org is the organization or username)
- <http://repo.org.pages.yourdomain.com:3000/file.html>
- <http://branch.repo.org.pages.yourdomain.com:3000/file.html>
- <http://org.pages.yourdomain.com:3000/> (if you have created a gitea-pages repo it'll be served on the root)

### Gitea config

There are multiple options to expose your repo's as a page, that you can use both at the same time.

- creating a gitea-pages repo with a gitea-pages branch and a gitea-pages topic
- adding a gitea-pages branch to any repo of choice and a gitea-pages topic
- adding a gitea-pages-allowall topic to your repo (easiest, but less secure)

#### gitea-pages repo

e.g. we'll use the `yourorg` org.

1. create a `gitea-pages` repo in `yourorg` org
2. Add a `gitea-pages` topic to this `gitea-pages` repo (this is used to opt-in your repo),
3. Create a `gitea-pages` branch in this `gitea-pages` repo.
4. Put your content in this branch. (eg file.html)

Your content will now be available on <http://yourorg.pages.yourdomain.com:3000/file.html>

#### any repo with configurable allowed branch/tag/commits

e.g. we'll use the `yourrepo` repo in the `yourorg` org and there is a `file.html` in the `master` branch and a `otherfile.html` in the `dev` branch. The `master` branch is your default branch.

1. Add a `gitea-pages` topic to the `yourrepo` repo (this is used to opt-in your repo).
2. Create a `gitea-pages` branch in this `yourrepo` repo.
3. Put a `gitea-pages.toml` file in this `gitea-pages` branch of `yourrepo` repo. (more info about the content below)

The `gitea-pages.toml` file will contain the git reference (branch/tag/commit) you allow to be exposed.
To allow everything use the example below:

```toml
allowedrefs=["*"]
```

To only allow main and dev:

```toml
allowedrefs=["main","dev"]
```

- Your `file.html` in the `master` branch will now be available on <http://yourorg.pages.yourdomain.com:3000/yourrepo/file.html>
- Your `file.html` in the `master` branch will now be available on <http://yourrepo.yourorg.pages.yourdomain.com:3000/file.html>
- Your `otherfile.html` in the `dev` branch will now be available on <http://yourorg.pages.yourdomain.com:3000/yourrepo/file.html?ref=dev>
- Your `otherfile.html` in the `dev` branch will now be available on <http://dev.yourrepo.yourorg.pages.yourdomain.com:3000/file.html>

#### any repo with all branches/tags/commits exposed

e.g. we'll use the `yourrepo` repo in the `yourorg` org and there is a `file.html` in the `master` branch and a `otherfile.html` in the `dev` branch. The `master` branch is your default branch.

1. Add a `gitea-pages-allowall` topic to the `yourrepo` repo (this is used to opt-in your repo).

- Your `file.html` in the `master` branch will now be available on <http://yourorg.pages.yourdomain.com:3000/yourrepo/file.html>
- Your `file.html` in the `master` branch will now be available on <http://yourrepo.yourorg.pages.yourdomain.com:3000/file.html>
- Your `otherfile.html` in the `dev` branch will now be available on <http://yourorg.pages.yourdomain.com:3000/yourrepo/file.html?ref=dev>
- Your `otherfile.html` in the `dev` branch will now be available on <http://dev.yourrepo.yourorg.pages.yourdomain.com:3000/file.html>

## Building caddy

As this is a 3rd party plugin you'll need to build caddy (or use the binaries).
To build with this plugin you'll need to have go1.19 installed.

```go
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest #this will install xcaddy in ~/go/bin
~/go/bin/xcaddy build --with github.com/42wim/caddy-gitea@v0.0.4
```
