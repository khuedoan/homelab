# Updating documentation (this website)

This project uses the [Diátaxis](https://diataxis.fr) technical documentation framework.
The website is generated using [mdBook](https://rust-lang.github.io/mdBook) and can be viewed at [homelab.khuedoan.com](https://homelab.khuedoan.com).

There are 4 main parts:

- [Getting started (tutorials)](https://diataxis.fr/tutorials): learning-oriented
- [Concepts (explanation)](https://diataxis.fr/explanation): understanding-oriented
- [How-to guides](https://diataxis.fr/how-to-guides): goal-oriented
- [Reference](https://diataxis.fr/reference): information-oriented

## Local development

To edit and view locally, run:

```sh
just docs
```

Then visit [localhost:8000](http://localhost:8000)

## Deployment

It's running on GitHub Pages under the `pages` branch
(so even if my homelab goes down everyone can still read the documentation).

To publish:

```sh
just publish
```
