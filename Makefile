
templ:
	templ generate --watch --proxy="http://localhost:4222" --proxybind="localhost" --proxyport="4000"
start:
	air

.PHONY: templ start
