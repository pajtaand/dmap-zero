"""
This Proof of Concept (PoC) module is a Python-based client that interacts with a specified Agent API.
It demonstrates various functionalities including webhook registration, endpoint listing, and periodic data pushing to both endpoints and a controller.
"""

import os
import requests
import threading
import time
import base64
import tempfile
from flask import Flask, request, jsonify
from werkzeug.serving import make_server

# Save all other env variables into a dictionary as config
config = {key: value for key, value in os.environ.items()}

# API base URL
BASE_URL = config.get('MODULE_API_BASE_URL')
MODULE_API_BASEAUTH_USER = config.get('MODULE_API_BASEAUTH_USER')
MODULE_API_BASEAUTH_PASS = config.get('MODULE_API_BASEAUTH_PASS')
MODULE_API_CERTIFICATE = config.get('MODULE_API_CERTIFICATE')
MODULE_GIVEN_PORT = config.get('MODULE_GIVEN_PORT')

# Authentication tuple for requests
auth = (MODULE_API_BASEAUTH_USER, MODULE_API_BASEAUTH_PASS)

# Flask app for webhook endpoints
app = Flask(__name__)

# Create a temporary file to store the certificate
cert_file = tempfile.NamedTemporaryFile(delete=False)
cert_file.write(base64.b64decode(MODULE_API_CERTIFICATE))
cert_file.close()

# Webhook handlers
@app.route('/webhook1', methods=['POST'])
def webhook1():
    data = request.json
    message = base64.b64decode(data["blob"].encode()).decode()
    print(f"Webhook 1 received message: {message}")
    return jsonify({"status": "success"}), 200

@app.route('/webhook2', methods=['POST'])
def webhook2():
    data = request.json
    message = base64.b64decode(data["blob"].encode()).decode()
    print(f"Webhook 2 received message: {message}")
    return jsonify({"status": "success"}), 200

@app.route('/webhook3', methods=['POST'])
def webhook3():
    data = request.json
    message = base64.b64decode(data["blob"].encode()).decode()
    print(f"Webhook 3 received message: {message}")
    return jsonify({"status": "success"}), 200

@app.route('/webhook4', methods=['POST'])
def webhook4():
    data = request.json
    message = base64.b64decode(data["blob"].encode()).decode()
    print(f"Webhook 4 received message: {message}")
    return jsonify({"status": "success"}), 200

# Function to start Flask server
def start_flask(host, port):
    server = make_server(host, port, app)
    server.serve_forever()

# Start Flask server in a separate thread
flask_thread = threading.Thread(target=start_flask, args=('0.0.0.0', int(MODULE_GIVEN_PORT)))
flask_thread.start()

# Helper function for making requests with certificate
def make_request(method, url, **kwargs):
    verify = False
    # verify certificate only if set in config (only for development!)
    if config.get("VERIFY_CERTIFICATE", False):
        verify = cert_file.name
    return requests.request(method, url, auth=auth, verify=verify, **kwargs)

# Register webhooks
def register_webhooks():
    webhook_configs = [
        ('/webhook1', 'CONTROLLER_DATA'),
        ('/webhook2', 'CONTROLLER_DATA'),
        ('/webhook3', 'ENDPOINT_DATA'),
        ('/webhook4', 'ENDPOINT_DATA')
    ]
    
    for url_path, event in webhook_configs:
        payload = {
            "urlPath": url_path,
            "event": event
        }
        response = make_request('POST', f"{BASE_URL}/webhook", json=payload)
        if response.status_code == 201:
            print(f"Webhook registered successfully: {url_path} for event {event}")
        else:
            print(f"Failed to register webhook: {url_path} for event {event}")

# Function to list endpoints
def list_endpoints():
    response = make_request('GET', f"{BASE_URL}/endpoint")
    if response.status_code == 200:
        endpoints = response.json()
        print("Endpoints:")
        for endpoint in endpoints:
            print(f"- {endpoint['id']}")
    else:
        print("Failed to list endpoints")

# Function to push message to endpoint
def push_to_endpoint(endpoint_id, message):
    headers = {'Content-Type': 'application/octet-stream'}
    response = make_request('POST', f"{BASE_URL}/endpoint/push?id={endpoint_id}", 
                            data=message.encode(), 
                            headers=headers)
    if response.status_code == 200:
        print(f"Message pushed to endpoint {endpoint_id}")
    else:
        print(f"Failed to push message to endpoint {endpoint_id}")

# Function to push message to controller
def push_to_controller(receiver_id, message):
    payload = {
        "receiverId": receiver_id,
        "blob": base64.b64encode(message.encode()).decode()
    }
    response = make_request('POST', f"{BASE_URL}/controller/push", json=payload)
    if response.status_code == 200:
        print(f"Message pushed to controller for receiver {receiver_id}")
    else:
        print(f"Failed to push message to controller for receiver {receiver_id}")

# Main loop
def main_loop():
    print(f"Starting application with config: {config}")

    register_webhooks()
    
    try:
        while True:
            list_endpoints()
            
            # Push message to all endpoints (including this)
            response = make_request('GET', f"{BASE_URL}/endpoint")
            if response.status_code == 200:
                endpoints = response.json()
                for endpoint in endpoints:
                    push_to_endpoint(endpoint['id'], 'hey there, this is module')
            
            # Push message to controller
            push_to_controller('controller', 'hi controller, this is module')
            
            time.sleep(15)
    except KeyboardInterrupt:
        print("Stopping the client...")
    finally:
        # Graceful shutdown
        print("Shutting down...")
        time.sleep(5)  # Wait for 5 seconds before exiting
        # Remove the temporary certificate file
        os.unlink(cert_file.name)

if __name__ == "__main__":
    main_loop()
