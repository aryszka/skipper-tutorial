# Skipper Tutorial

Intro, a simple web document editor composed of multiple services, involving a viewer, an editor and an updater

## Step 0 - Starting Skipper without routes

```
skipper &
curl -v [::1]:9090
```

...or check [::1]:9090 in the browser. will result in 404, because skipper alone doesn't do anything

Skipper prints access logs by default. To prevent this for the time of the tutorial, we can silent it for all
subsequent starts:

```
alias skipper="skipper -access-log-disabled"
```


## Step 1 - Create a 404 page

This tutorial will contain some routes with certain constraints, 404 for catchall.

Create a directory, called static-content, and an html file, called static-content/404.html:

```
<!doctype html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <title>Skipper Tutorial - Page Not Found</title>
        <style>
			* {font-family: sans-serif}
        </style>
    </head>
    <body>
        <h1>Skipper Tutorial</h1>
        <p>The requested page cannot be found.</p>
    </body>
</html>
```

Create a routing file, called routes.eskip:

```
// 404 as catchall:
notFound: * -> status(404) -> modPath(".*", "/404.html") -> static("", "static-content") -> <shunt>;
```

Test this route by restarting skipper with new routes file and making an http request:

```
pkill skipper; skipper -routes-file routes.eskip &
curl [::1]:9090
```

Check the 404 page in the browser.

How this route works: * means all requests are handled by this route that don't match routes with more specific
constraints (predicates). For the request, the filter status(404) does nothing, modPath changes any incoming path
(.*) to /404.html, and static("", "static-content") serves the request from the static-content directory, with
the relative path specified by request, overwritten in the modPath filter. <shunt> does nothing, it means that
the request is not forwarded to a backend service. When the static filter is done, 

One possible excercise would be here to create an additional 404 route that matches those requests, that need to
be served 404, but do not accept text/html, using the Header or HeaderRegexp predicates.

The reason for the position of the status filter.


## Step 2 - Create a viewer service

Create a directory, called view, and an html template, called view.html.

```
<!doctype html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <title>Skipper Tutorial - Page Not Found</title>
        <style>
			* {font-family: sans-serif}
        </style>
    </head>
    <body>
        <h1>Skipper Tutorial</h1>
        <h3>{{.Title}}</h3>
        <code>
            {{.Content}}
        </code>
    </body>
</html>
```

In the same directory creat a go file, called view.go.

```
package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "strings"
    "path"
)

var page *template.Template

func init() {
	t, err := template.ParseFiles("view/view.html")
	if err != nil {
		panic(err)
	}

	page = t
}

func handle(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(path.Join("data", p))
    if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
        return
    }

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

    if strings.HasPrefix(r.Header.Get("Accept"), "text/plain") {
        w.Write(content)
        return
    }

	page.Execute(w, string(content))
}

func main() {
	log.Fatal(http.ListenAndServe(":9091", http.HandlerFunc(handle)))
}
```

This is a service that takes a file from the data directory, defined by the request path, and renders its
content as HTML. As an additional feature, when the request accepts only plain text, it returns the content of
the file as it is.

This is a standalone service, it can be tested separately, without Skipper. To display some content, create a
directory, called data and create a file in it, called hello:

```
mkdir data && echo Hello, world! > data/hello
```

To test the standalone service, start it and call it with the path /hello:

```
go run view/view.go &
curl -H Accept:\ text/plain [::1]:9091/hello
```

To see it as html, open the same url in the browser.

Any route that doesn't match still goes to the 404 page.

Now, to access this service from a common HTTP address, we can add a new route to the existing 404 route. We
can define separate routes for the text and the HTML representations. Extend routes.eskip by adding the text and
html routes:

```
// 404 as catchall:
notFound: * -> status(404) -> modPath(".*", "/404.html") -> static("", "static-content") -> <shunt>;

// View content as plain:
plain:
	Path("/content/:fn/plain")
	-> modPath("^/content/(?P<fn>[^/]+)/plain$", "/$fn")
	-> setRequestHeader("Accept", "text/plain")
	-> "http://[::1]:9091";

// ...or as html:
html:
	Path("/content/:fn/html")
	-> modPath("^/content/(?P<fn>[^/]+)/html$", "/$fn")
	-> "http://[::1]:9091";
```

Restart skipper and test the new routes:

```
pkill skipper && skipper -routes-file routes.eskip &
curl [::1]:9090/content/hello/plain
curl [::1]:9090/content/hello/html
```

What these routes do.

Fix for the regexp stuff.

```
// 404 as catchall:
notFound: * -> status(404) -> modPath(".*", "/404.html") -> static("", "static-content") -> <shunt>;

// View content as plain:
plain:
	Path("/content/:fn/plain")
	-> modPath(".*", "/{{.fn}}")
	-> setRequestHeader("Accept", "text/plain")
	-> "http://[::1]:9091";

// ...or as html:
html:
	Path("/content/:fn/html")
	-> modPath(".*", "/{{.fn}}")
	-> "http://[::1]:9091";
```

We can use the params also to use a single route:

```
// 404 as catchall:
notFound: * -> status(404) -> modPath(".*", "/404.html") -> static("", "static-content") -> <shunt>;

// View content as plain or html:
view:
	Path("/content/:fn/:format")
	-> modPath(".*", "/{{.fn}}")
	-> setRequestHeader("Accept", "text/{{.format}}")
	-> "http://[::1]:9091";
```


## Step 3 - Create an updater service

Accept as form encoded, write to the file.

Make a directory called update, and a go file called update/update.go:

```
package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"io"
)

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := os.Create(path.Join("data", r.URL.Path))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.Copy(f, r.Body)
}

func main() {
	log.Fatal(http.ListenAndServe(":9092", http.HandlerFunc(handle)))
}
```

The updater is also a standalone service, and can be tested:

```
go run update/update.go &
curl -X PUT -d 'Hello, hello, hello!' [::1]:9092/hello
curl [::1]:9090/content/hello/plain
```

Add a route, extend the `routes.eskip` with this:

```
// Update content:
update:
	Method("PUT") && Path("/content/:fn")
	-> modPath(".*", "/{{.fn}}")
	-> "http://localhost:9092";
```

Test:

```
pkill skipper; skipper -routes-file routes.eskip &
curl -X PUT -d 'Hello, hello, hello, again!' [::1]:9090/content/hello
curl [::1]:9090/content/hello/plain
```

## Step 4 - Avoid restarts, use etcd

Installing and starting etcd is easy:

```
go get github.com/coreos/etcd/...
etcd 2> /dev/null &
```

Telling skipper to use etcd instead of the file can be done by restarting with the etcd argument:

```
pkill skipper; skipper -etcd-urls http://[::1]:2379 &
```

```
eskip upsert -routes 'example: Path("/get") -> "https://httpbin.org"'
```

Skipper should receive the updates in a few seconds, and then we can test the example route:

```
curl [::1]:9090/get
```

To delete the example route, we can use eskip:

```
eskip delete -ids example
```

Skipper comes with a supplementary tool: eskip. It can be used to maintain routes in etcd.

```
eskip upsert routes.eskip
```

Now our routes should work again, but now loaded from etcd. To test:


## Step 5 - Content editor

We create an editor. Loaded as HTML, JS loads and updates the content.

Create a file in the static-content directory, called edit.html:

```
```


## Step 6 - Protect the services


## Step 7 - Test a new version of the viewer


## Step 8 - Test the new version with traffic control


## Step 9 - Apply the new version of the viewer
