# PoC Module


## Start Module Controller
```bash
python3 poc-module/webhook-handler-script.py 974e7e55-5115-4588-8878-a2d76ea6b5f4 --username username --password password --api_url https://192.168.0.5:6969/api/v1 --advertised_address "$(hostname -I | awk '{print $1}').sslip.io" --advertised_port 3104
```
