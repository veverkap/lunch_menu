#!/usr/bin/python
import requests
from bs4 import BeautifulSoup
import hashlib
import os

url = 'https://virginiabeach.revtrak.net/rw-class-registrations/'
response = requests.get(url)

if response.status_code == 200:
    soup = BeautifulSoup(response.text, 'html.parser')
    content = soup.prettify()

    # Calculate the hash of the page content
    content_hash = hashlib.sha256(content.encode('utf-8')).hexdigest()

    # Read the previous hash from the file, if it exists
    hash_file = 'page_hash.txt'
    if os.path.exists(hash_file):
        with open(hash_file, 'r') as file:
            previous_hash = file.read().strip()
    else:
        previous_hash = ''

    # Compare the current hash with the previous hash
    if content_hash != previous_hash:
        # Update the hash file with the new hash
        with open(hash_file, 'w') as file:
            file.write(content_hash)

        # Set the environment variable to indicate changes
        print(f"::set-env name=has_changes::true")
    else:
        print("No changes detected.")
else:
    print(f"Failed to fetch the page. Status code: {response.status_code}")
