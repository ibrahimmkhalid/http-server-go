## HTTP Server implementation in Go

Based on the
["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview).

[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the
protocol that powers the web. In this challenge, I built a HTTP/1.1 server
that is capable of serving multiple clients.

Along the way I learned about TCP servers,
[HTTP request syntax](https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html),
and more.

---

Endpoints developed:

- `/` - Responds with a simple 200 HTTP code
- `/echo/...` - Responds with a simple text echo of anything after the `/echo/`
- `/user-agent` - Responds with a simple text of the user agent that sent the `GET` request
- `GET /files/...` - Responds with the contents of the file specified if it exists, otherwise a 404 error
- `POST /files/...` - Responds with a 201 Created HTTP code after saving the contents of the request body to the specified file

---

**Note**: If you're viewing this repo on GitHub, head over to
[codecrafters.io](https://codecrafters.io) to try the challenge.

1. Ensure you have `go (1.19)` installed locally
1. Run `./your_server.sh` to run your program, which is implemented in
   `app/server.go`.
1. Commit your changes and run `git push origin master` to submit your solution
   to CodeCrafters. Test output will be streamed to your terminal.
