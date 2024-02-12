# PassThru.fun

Passthru is an open-source HTTP tunneling tool, similar to ngrok, but fully open-source. It allows you to expose a web server running on your local machine to the internet. Just tell Passthru what port your web server is listening on, and it does the rest!



## Running as a Client

To run Passthru as a client, use the `-type` flag and set it to `client` (default is client). You must also specify a domain using the `-domain` flag. This will be used as a subdomain. For example, if you specify `mydomain` as the domain, then your local port will be available via `mydomain.passthru.fun`.

Example:
```bash
./passthru -type client -domain mydomain -port 8888
```

You can also enable verbose mode by using the `-verbose` flag.

Example:
```bash
./passthru -type client -domain mydomain -port 8888 -verbose true
```




## Running as a Server

To run Passthru on a server, use the `-type` flag and set it to `server`. You can also specify the port to run on using the `-port` flag.

Example:
```bash
./passthru -type server -port 8888
```
## License

This project is licensed under the MIT License - the most permissive and open-source license available.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any problems or have any questions, please open an issue on GitHub.