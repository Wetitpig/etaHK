use std::string;
use std::thread;
use std::time::Duration;

use reqwest::{Url,StatusCode,Client};
use serde_json::Value;
use serde_json::json;
async fn emitRequest(url: string::String) -> string::String {
	let client = Client::new();
	loop {
		let endpoint = Url::parse(url.as_str()).unwrap();
		let response = client.get(endpoint).send().await.unwrap();
		match response.status() {
			StatusCode::OK => {
				return response.text().await.unwrap();
			}
			StatusCode::TOO_MANY_REQUESTS => thread::sleep(Duration::from_secs(1)),
			s => panic!("Error on request to {}: {}", url, s)
		}
	}
}

pub async fn getRouteStop(co: string::String) -> (serde_json::Value, serde_json::Value) {
	let stop_list = json!({});

	let mut r: Value = serde_json::from_str(
		emitRequest(String::from("https://rt.data.gov.hk/v1.1/transport/citybus-nwfb/route/") + co.as_str()).await.as_str()
	).unwrap();
	let route_list = r["data"].take();
	return (route_list, stop_list);
}

