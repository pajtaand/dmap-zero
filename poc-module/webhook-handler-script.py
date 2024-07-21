"""
Webhook Handler and Data Sender Script

Interacts with a agent controller API to register a webhook, handle incoming webhook data, and periodically send data to a specified module.

Usage:
    python3 webhook-handler-script.py <module_id> [--api_url API_URL]

Arguments:
    module_id : str
        The ID of the module to communicate with
    --api_url : str, optional
        The base URL of the API
    --advertised_address : str, optional
        The advertised address of the host (e.g. ip or hostname)
    --username : str, optional
        Username for basic authentication
    --password : str, optional
        Password for basic authentication

Note: SSL certificate verification is disabled for development use.
"""

import argparse
import json
import logging
import sys
import time
import base64
import threading
from http.server import HTTPServer, BaseHTTPRequestHandler
import requests
from requests.auth import HTTPBasicAuth

# Constants
DEFAULT_API_URL = "https://localhost:6969/api/v1"
DEFAULT_WEBHOOK_PORT = 3358
SEND_INTERVAL = 15

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class WebhookHandler:
    def __init__(self, module_id, api_url, advertised_address, advertised_port, username, password):
        self.module_id = module_id
        self.api_url = api_url
        self.advertised_address = advertised_address
        self.advertised_port = advertised_port
        self.auth = HTTPBasicAuth(username, password) if username and password else None
        self.webhook_id = None
        self.stop_event = threading.Event()

    def register_webhook(self):
        webhook_data = {
            "moduleID": self.module_id,
            "url": f"http://{self.advertised_address}:{self.advertised_port}/webhook",
        }
        response = requests.post(f"{self.api_url}/webhook", json=webhook_data, auth=self.auth, verify=False)
        if response.status_code == 201:
            response_data = response.json()
            self.webhook_id = response_data.get('ID')
            if self.webhook_id:
                logger.info(f"Webhook registered successfully with ID: {self.webhook_id}")
            else:
                logger.error("Webhook ID not found in the response")
                sys.exit(1)
        else:
            logger.error(f"Failed to register webhook: {response.status_code}")
            sys.exit(1)

    def delete_webhook(self):
        if self.webhook_id:
            response = requests.delete(f"{self.api_url}/webhook?id={self.webhook_id}", auth=self.auth, verify=False)
            if response.status_code == 204:
                logger.info(f"Webhook {self.webhook_id} deleted successfully")
            else:
                logger.error(f"Failed to delete webhook {self.webhook_id}: {response.status_code}")

    def send_data(self):
        while not self.stop_event.is_set():
            message = "hello world, I'm from outside"
            data = {"data": base64.b64encode(message.encode()).decode()}
            response = requests.post(f"{self.api_url}/module/{self.module_id}/send", json=data, auth=self.auth, verify=False)
            if response.status_code == 200:
                logger.info(f"Data sent successfully to module {self.module_id}")
            else:
                logger.error(f"Failed to send data: {response.status_code}")
            time.sleep(SEND_INTERVAL)

    class WebhookRequestHandler(BaseHTTPRequestHandler):
        def do_POST(self):
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            data = json.loads(post_data.decode('utf-8'))
            logger.info(f"Received data: {json.dumps(data, indent=2)}")
            logger.info(f"Message: {base64.b64decode(data['blob'].encode()).decode()}")
            self.send_response(200)
            self.end_headers()

    def run_webhook_server(self):
        server = HTTPServer((self.advertised_address, int(self.advertised_port)), self.WebhookRequestHandler)
        logger.info(f"Webhook server running on http://{self.advertised_address}:{self.advertised_port}")
        while not self.stop_event.is_set():
            server.handle_request()

    def run(self):
        self.register_webhook()

        webhook_thread = threading.Thread(target=self.run_webhook_server)
        webhook_thread.start()

        send_thread = threading.Thread(target=self.send_data)
        send_thread.start()

        # Keep the main thread alive
        while True:
            time.sleep(0.1)  # Sleep for an hour (or any long duration)

def parse_arguments():
    parser = argparse.ArgumentParser(description="Webhook handler and data sender")
    parser.add_argument("module_id", help="Module ID to communicate with")
    parser.add_argument("--api_url", default=DEFAULT_API_URL, help="API URL (default: %(default)s)")
    parser.add_argument("--advertised_address", help="The advertised address of the host (e.g. ip or hostname)")
    parser.add_argument("--advertised_port", default=DEFAULT_WEBHOOK_PORT, help="The advertised port of the host (e.g. 3456)")
    parser.add_argument("--username", help="Username for basic authentication")
    parser.add_argument("--password", help="Password for basic authentication")
    return parser.parse_args()

if __name__ == "__main__":
    args = parse_arguments()
    handler = WebhookHandler(args.module_id, args.api_url, args.advertised_address, args.advertised_port, args.username, args.password)
    handler.run()
