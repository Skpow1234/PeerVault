#!/usr/bin/env python3
"""
PeerVault REST API - Basic File Upload Example

This example demonstrates how to upload a file to PeerVault using the REST API.
"""

import requests
import json
import os
import sys
from pathlib import Path

# Configuration
PEERVAULT_BASE_URL = "http://localhost:8080"
API_BASE_URL = f"{PEERVAULT_BASE_URL}/api/v1"

def upload_file(file_path, key=None):
    """
    Upload a file to PeerVault.
    
    Args:
        file_path (str): Path to the file to upload
        key (str, optional): Key to store the file under. If not provided, uses filename.
    
    Returns:
        dict: Upload response from the API
    """
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"File not found: {file_path}")
    
    if key is None:
        key = os.path.basename(file_path)
    
    # Prepare the file for upload
    with open(file_path, 'rb') as file:
        files = {
            'file': (os.path.basename(file_path), file, 'application/octet-stream')
        }
        data = {
            'key': key
        }
        
        # Make the upload request
        response = requests.post(
            f"{API_BASE_URL}/files",
            files=files,
            data=data,
            timeout=30
        )
    
    # Check for errors
    response.raise_for_status()
    
    return response.json()

def download_file(key, output_path=None):
    """
    Download a file from PeerVault.
    
    Args:
        key (str): Key of the file to download
        output_path (str, optional): Path to save the file. If not provided, uses the key.
    
    Returns:
        str: Path where the file was saved
    """
    if output_path is None:
        output_path = key
    
    # Make the download request
    response = requests.get(
        f"{API_BASE_URL}/files/{key}",
        timeout=30
    )
    
    # Check for errors
    response.raise_for_status()
    
    # Save the file
    with open(output_path, 'wb') as file:
        file.write(response.content)
    
    return output_path

def list_files():
    """
    List all files in PeerVault.
    
    Returns:
        list: List of file metadata
    """
    response = requests.get(f"{API_BASE_URL}/files", timeout=30)
    response.raise_for_status()
    return response.json()

def delete_file(key):
    """
    Delete a file from PeerVault.
    
    Args:
        key (str): Key of the file to delete
    
    Returns:
        dict: Delete response from the API
    """
    response = requests.delete(f"{API_BASE_URL}/files/{key}", timeout=30)
    response.raise_for_status()
    return response.json()

def get_file_info(key):
    """
    Get information about a specific file.
    
    Args:
        key (str): Key of the file
    
    Returns:
        dict: File metadata
    """
    response = requests.get(f"{API_BASE_URL}/files/{key}/info", timeout=30)
    response.raise_for_status()
    return response.json()

def main():
    """Main function demonstrating basic file operations."""
    print("PeerVault REST API - Basic File Operations Example")
    print("=" * 50)
    
    # Create a sample file
    sample_file = "sample.txt"
    sample_content = "Hello, PeerVault! This is a sample file for testing."
    
    with open(sample_file, 'w') as f:
        f.write(sample_content)
    
    try:
        # Upload the file
        print(f"Uploading file: {sample_file}")
        upload_result = upload_file(sample_file, "my-sample-file")
        print(f"Upload successful: {json.dumps(upload_result, indent=2)}")
        
        # Get file info
        print(f"\nGetting file info for: my-sample-file")
        file_info = get_file_info("my-sample-file")
        print(f"File info: {json.dumps(file_info, indent=2)}")
        
        # List all files
        print(f"\nListing all files:")
        files = list_files()
        for file in files:
            print(f"  - {file['key']} ({file['size']} bytes)")
        
        # Download the file
        print(f"\nDownloading file: my-sample-file")
        downloaded_path = download_file("my-sample-file", "downloaded-sample.txt")
        print(f"Downloaded to: {downloaded_path}")
        
        # Verify the content
        with open(downloaded_path, 'r') as f:
            downloaded_content = f.read()
        
        if downloaded_content == sample_content:
            print("✅ Content verification successful!")
        else:
            print("❌ Content verification failed!")
        
        # Delete the file
        print(f"\nDeleting file: my-sample-file")
        delete_result = delete_file("my-sample-file")
        print(f"Delete successful: {json.dumps(delete_result, indent=2)}")
        
        # List files again to confirm deletion
        print(f"\nListing files after deletion:")
        files = list_files()
        if files:
            for file in files:
                print(f"  - {file['key']} ({file['size']} bytes)")
        else:
            print("  No files found")
        
    except requests.exceptions.RequestException as e:
        print(f"❌ API request failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response status: {e.response.status_code}")
            print(f"Response body: {e.response.text}")
    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        # Clean up sample files
        for file in [sample_file, "downloaded-sample.txt"]:
            if os.path.exists(file):
                os.remove(file)
                print(f"Cleaned up: {file}")

if __name__ == "__main__":
    main()
