# Envoy Discovery Service

This is a proof of concept implementation of a management server for Enovy to communicate dynamically with.

## Overview

The XDS will listen on http://127.0.0.1:19000 for gRPC requests from Envoy, and prepopulate a cluster based on the [xds.yaml](configs/xds.yaml)

XDS will then expose an HTTP server on http://127.0.0.1:19001 that will look for a JSON POST with a path and a port:

```
{
"path": "/path",
"port": "9001"
}
```
## Usage

Start the XDS server:

```
$ ./builds/envoy-xds-darwin
```

In another terminal, start Envoy:

```
$ envoy -c configs/xds.yaml
```

In yet another terminal, post to the HTTP server:

```
curl -X POST http://localhost:19001/addPath -d @configs/addPath.json
```

You will see similar output to this from Envoy:

```
...
[2018-06-07 20:12:49.013][13383733][info][upstream] source/common/upstream/cluster_manager_impl.cc:379] add/update cluster nick-xds during init
[2018-06-07 20:12:49.013][13383733][info][upstream] source/common/upstream/cluster_manager_impl.cc:131] cm init: all clusters initialized
[2018-06-07 20:12:49.013][13383733][info][main] source/server/server.cc:386] all clusters initialized. initializing init manager
[2018-06-07 20:12:49.022][13383733][info][upstream] source/server/lds_api.cc:61] lds: add/update listener 'listener-0'
[2018-06-07 20:12:49.022][13383733][info][config] source/server/listener_manager_impl.cc:607] all dependencies initialized. starting workers
[2018-06-07 20:13:17.912][13383733][info][upstream] source/common/upstream/cluster_manager_impl.cc:385] add/update cluster port9001 starting warming
[2018-06-07 20:13:17.912][13383733][info][upstream] source/common/upstream/cluster_manager_impl.cc:392] warming cluster port9001 complete
[2018-06-07 20:13:17.920][13383733][info][upstream] source/server/lds_api.cc:61] lds: add/update listener 'listener-0'
```
