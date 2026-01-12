# Go-Bridge

A demonstration of reverse proxying from scratch, using only the standard Golang library.

Project structure built with [Gini](https://github.com/rafaeldepontes/goinit)

## How to use

You need two terminals...

first clone the repo:

```bash
git clone <repo>
cd gobridge
```

On the first terminal run:

```bash
go run .
```

And on the second one runthis simple curl:

```bash
curl http://localhost:8080/todos/{any id up to 200}
```
