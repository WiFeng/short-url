# short-url

## Rest api

* admin/create
  
    ```shell
        curl --location --request POST 'http://127.0.0.1:8081/admin/create' \
            --header 'Content-Type: text/plain' \
            --data-raw '{
                "long_url" : "https://github.com/wifeng/leetcode"
            }'
    ```

    ```shell
        {
            "short_url": "http://sh.url/2bI"
        }
    ```

* admin/query

    ```shell
        curl --location --request POST 'http://127.0.0.1:8081/admin/query' \
            --header 'Content-Type: text/plain' \
            --data-raw '{
                "short_url" : "2bI"
            }'
    ```

    ```shell
        {
            "long_url": "https://github.com/wifeng/leetcode"
        }
    ```

## Redirect

Request

```shell
    http://sh.url/2bI
```

Response

```shell
    HTTP/1.1 302 Found
    Content-Type: application/json; charset=utf-8
    Location: https://github.com/wifeng/leetcode
    Date: Mon, 01 Jun 2020 05:28:03 GMT
    Content-Length: 0
```
