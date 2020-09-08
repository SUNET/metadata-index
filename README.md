Datasets index service
=====

API
---

register
--------

```
POST /api/register
<url>
```

Register a URL with the index server. The URL will be dereferenced and if it contains valid JSON-LD fulfilling the manifests schema the manifest metadata (eg creator, publisher, rightsHolder elements) will be indexed along the @id. The assumption is that the @id URL can be dereferenced in the future to update the index.

*Example*

```
# curl -XPOST --data "https://raw.githubusercontent.com/SUNET/metadata-manifests/master/examples/example.jsonld" http://localhost:3000/api/register
```

search
------

```
GET /api/search?q=<query string>
```

Return a JSON list of @id elements. If q is absent return all elements in the index.

*Eample*

```
# curl http://localhost:3000/api/search?q=example.edu
```

status
------

```
GET /api/status
```

Return a status JSON document containing the sw version and the number of documents in the index.
