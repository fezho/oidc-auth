# OpenID Connect Authentication Service 

This is an HTTP Server that offers authentication and session management with [OpenID Connect](https://openid.net/connect/). 

As an OpenID Connect client, our auth service can easily verify the identity of an end user by connecting to an authorization server or identity provider (IdP), as well as to obtain basic profile information about the end user.

## OpenID Connect

OpenID Connect is a simple identity layer on top of the OAuth 2.0 protocol which allows a range of clients, including web-based, mobile, and JavaScript clients, to request and receive information about authenticated sessions and end-users. 

### Why OpenID Connect

OpenID Connect is easily integrated, and it can work with a wider variety of apps. Specifically, it provides:
*  Easily consumed identity tokens

   Client apps receive the user’s identity encoded in a secure JSON Web Token (JWT) called the ID token.
  
*  The OAuth 2.0 protocol

   Clients use OAuth 2.0 flows to obtain ID tokens, which work with web apps as well as native mobile apps.


*  Simplicity with capability

   OpenID Connect is simple enough to integrate with basic apps.
   
### Grant Types

// https://developers.onelogin.com/openid-connect
Our service supports the following two OpenID Connect authentication flows:

*  **Implicit Flow** is required for apps that have no “back end” logic on the web server, like a Javascript app.
*  **Authentication (or Basic) Flow** is designed for apps that have a back end that can communicate with the IdP away from prying eyes.

Currently our service only supports OIDC's Authorization Code Flow


// TODO: use envoy auth proxy as example:
https://serialized.net/2019/05/envoy-ratelimits/
https://github.com/jbarratt/envoy_ratelimit_example

https://www.getambassador.io/docs/latest/howtos/basic-auth/

