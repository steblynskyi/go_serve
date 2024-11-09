# GO-Serve

Go-Serve it is minimalistic web server which will serve react bundle static files and will expose endpoint to get configuration on browser rendering (SHOULD NOT HAVE SECRETS).

Config endpoint will be based on app type if app is running under L7 rule with url segment like: `domain.com/namespace/*` then path to config will be `domain.com/namespace/env`.json if app running on root path `domain.com/env.json`. Also env config can be loaded as js (window object) file from `domain.com/env.js` or `domain.com/namespace/env.js`.

***Assets url in index file must point to exact path if L7 used `namespace/css,js`***

There are command line options to run go-serve:
```
  -address string
    	address to listen on (default ":8080")
  -base-path string
    	base url path to serve
  -dir string
    	bundle directory to serve
  -env-prefix string
    	environment variables prefix to look (default "APP_")
  -help
    	print help
  -read-timeout int
    	read timeout in seconds (default 15)
  -version
    	print version
  -write-timeout int
    	write timeout in seconds (default 15)
  -security-content-security-policy string
    	ContentSecurityPolicy sets the `Content-Security-Policy` header providing security against cross-site scripting (XSS), clickjacking and other code injection attacks (default "")
  -security-xss-protection string
    	XSSProtection provides protection against cross-site scripting attack (XSS)(default "1; mode=block")
  -security-contenttype-nosniff string
    	ContentTypeNosniff provides protection against overriding Content-Type (default "nosniff")
  -security-xframe-options string
    	XFrameOptions can be used to indicate whether or not a browser should be allowed to render a page in a <frame>, <iframe> or <object>(default "SAMEORIGIN")
  -security-hsts-maxage int
    	HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how long (in seconds) browsers should remember that this site is only to be accessed using HTTPS.(default 0)
  -security-hsts-excluede-subdomains bool
    	HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`header (default false)
  -security-disable bool
    	Disable security middleware (security-content-security-policy, security-xss-protection, security-contenttype-nosniff, security-xframe-options, security-hsts-maxage,security-hsts-excluede-subdomains ) (default false)
  -set-custom-header string
    	set custom response header format: 'HEADER_NAME:HEADER_VALUE'
```

Examples:

  ```
  go-serve -dir ./bundle -env-prefix GO_SERVE_ -address :8080 -base-path /go-serve/ -read-timeout 15 -write-timeout 15
  ```
  ```
  go run main.go --security-content-security-policy "default-src 'self'"
  ```


***Configuration for app*** should be passed as environment variables. And it will only be available on browser rendering if environment variables prefix matches with what is passed to -env-prefix option.


# GO-Serve Code Changes

-  Update public path in webpack.config.js to reflect load balancer URL. example : Config web UI load at app.dev-steblynskyi.com/admin/config/. so the public path should be  : publicPath: isLocal ? '/' : '/admin/config/ [https://bitbucket.org/steblynskyi/steblynskyi.bookingengine.config.web/src/master/webpack.config.js#lines-88] ( https://bitbucket.org/steblynskyi/steblynskyi.bookingengine.config.web/src/master/webpack.config.js#lines-88 )

- Use @steblynskyi-react-core/config package to configure environment variables in service. This package make an API call to - BASE_PATH/env.json & in response you will receive the env variables configured in container. [https://bitbucket.org/steblynskyi/steblynskyi.bookingengine.config.web/src/master/src/components/AppLayout/AppLayout.js#lines-3] ( https://bitbucket.org/steblynskyi/steblynskyi.bookingengine.config.web/src/master/src/components/AppLayout/AppLayout.js#lines-3)

- You will find all configured environment variables with prefix APP_. [https://ops-artifactory-manager.dev-steblynskyi.com/k8s/kube.dev-steblynskyi.com/be/booking-engine-checkin-web] (https://ops-artifactory-manager.dev-steblynskyi.com/k8s/kube.dev-steblynskyi.com/be/booking-engine-checkin-web ) However, when you make /env.json api call, you will get env variables key values without APP_ prefix.  ( APP_ prefix is needed to differenciate env variables app using vs env variables coming up with container)

- Setting Custom headers

 go run main.go -custom-header "content-security-policy-report-only:default-src 'self';object-src 'none'; worker-src 'none'; base-uri https://app.dev-steblynskyi.com;" -custom-header "header2:value2"

go run main.go -custom-header "content-security-policy-report-only:default-src 'self'" -custom-header "content-security-policy-report-only:object-src 'none'; worker-src 'none'; base-uri https://app.dev-steblynskyi.com;" -custom-header "header2:value2"

Note : If you pass same key in -custom-header it will appened with existing header with ";" separator


