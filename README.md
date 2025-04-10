# Api

This is a collection of tools for building APIs in go. There are two primary integrations.

- proxy
- rpc

These setups abuse Go's quirk where adding a handler to something that ends with a `/` such as `rpc/` will match
anything that under that path as well. So `rpc/anything` will match.

## Proxy

This is a fairly simple proxy with an ability to add stacked middleware. You don't get access to the body in the
middleware, just the ability to manipulate headers.

## RPC

This is a more complicated RPC setup for RPC style function calling. It also operates on stacked middleware, but you can
get access to the raw bytes, you can't change them though.

RPC has an additional feature where a call to `GET /_info` will respond with integration request info. This is for
working with codegen on clients.
