package main

import (
	"fmt"
	"io"
	"net/http"
)

func main2() {

	url := "https://pro-api.coinmarketcap.com/v3/fear-and-greed/latest"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-CMC_PRO_API_KEY", "9af6a5f1eca94e65a4ea29947bbdad00")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

	/** body内容为，但是只保存data.value就行。
		{
	  "data": {
	    "value": 40,
	    "value_classification": "Neutral",
	    "update_time": "2024-09-19T02:54:56.017Z"
	  },
	  "status": {
	    "timestamp": "2026-03-05T22:43:48.471Z",
	    "error_code": 0,
	    "error_message": "",
	    "elapsed": 10,
	    "credit_count": 1,
	    "notice": ""
	  }
	}
	*/
}
