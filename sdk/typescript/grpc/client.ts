import { grpc } from '@improbable-eng/grpc-web';
import { PeerVaultServiceClient } from './generated/peervault_grpc_web_pb';
import {
  FileRequest,
  FileResponse,
  ListFilesRequest,
  ListFilesResponse,
  DeleteFileResponse,
  UpdateFileMetadataRequest,
  AddPeerRequest,
  PeerRequest,
  PeerResponse,
  ListPeersResponse,
  RemovePeerResponse,
  PeerHealthResponse,
  SystemInfoResponse,
  MetricsResponse,
  HealthResponse,
  FileOperationEvent,
  PeerEvent,
  SystemEvent,
  FileChunk,
} from './generated/peervault_pb';

export interface PeerVaultConfig {
  endpoint: string;
  authToken: string;
  timeout?: number;
  retryAttempts?: number;
}

export class PeerVaultClient {
  private client: PeerVaultServiceClient;
  private config: PeerVaultConfig;

  constructor(config: PeerVaultConfig) {
    this.config = {
      timeout: 30000,
      retryAttempts: 3,
      ...config,
    };

    this.client = new PeerVaultServiceClient(this.config.endpoint, {
      transport: grpc.WebsocketTransport(),
      debug: false,
    });
  }

  // File Operations

  async uploadFile(fileKey: string, fileData: Uint8Array): Promise<FileResponse> {
    return new Promise((resolve, reject) => {
      const stream = this.client.uploadFile();
      const chunkSize = 64 * 1024; // 64KB chunks
      let offset = 0;

      stream.on('data', (response: FileResponse) => {
        resolve(response);
      });

      stream.on('error', (error: grpc.Error) => {
        reject(new Error(`Upload failed: ${error.message}`));
      });

      // Send file in chunks
      const sendChunk = () => {
        if (offset >= fileData.length) {
          stream.end();
          return;
        }

        const end = Math.min(offset + chunkSize, fileData.length);
        const chunk = new FileChunk();
        chunk.setFileKey(fileKey);
        chunk.setData(fileData.slice(offset, end));
        chunk.setOffset(offset);
        chunk.setIsLast(end === fileData.length);
        chunk.setChecksum(this.calculateChecksum(fileData.slice(offset, end)));

        stream.write(chunk);
        offset = end;

        // Continue sending chunks
        setTimeout(sendChunk, 0);
      };

      sendChunk();
    });
  }

  async downloadFile(fileKey: string): Promise<Uint8Array> {
    return new Promise((resolve, reject) => {
      const request = new FileRequest();
      request.setKey(fileKey);

      const stream = this.client.downloadFile(request);
      const chunks: Uint8Array[] = [];

      stream.on('data', (chunk: FileChunk) => {
        chunks.push(chunk.getData_asU8());
      });

      stream.on('end', () => {
        // Combine all chunks
        const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
        const result = new Uint8Array(totalLength);
        let offset = 0;

        for (const chunk of chunks) {
          result.set(chunk, offset);
          offset += chunk.length;
        }

        resolve(result);
      });

      stream.on('error', (error: grpc.Error) => {
        reject(new Error(`Download failed: ${error.message}`));
      });
    });
  }

  async listFiles(page: number = 1, pageSize: number = 10, filter: string = ''): Promise<ListFilesResponse> {
    return new Promise((resolve, reject) => {
      const request = new ListFilesRequest();
      request.setPage(page);
      request.setPageSize(pageSize);
      request.setFilter(filter);

      this.client.listFiles(request, this.getMetadata(), (error: grpc.Error, response: ListFilesResponse) => {
        if (error) {
          reject(new Error(`List files failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async getFile(fileKey: string): Promise<FileResponse> {
    return new Promise((resolve, reject) => {
      const request = new FileRequest();
      request.setKey(fileKey);

      this.client.getFile(request, this.getMetadata(), (error: grpc.Error, response: FileResponse) => {
        if (error) {
          reject(new Error(`Get file failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async deleteFile(fileKey: string): Promise<DeleteFileResponse> {
    return new Promise((resolve, reject) => {
      const request = new FileRequest();
      request.setKey(fileKey);

      this.client.deleteFile(request, this.getMetadata(), (error: grpc.Error, response: DeleteFileResponse) => {
        if (error) {
          reject(new Error(`Delete file failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async updateFileMetadata(fileKey: string, metadata: Record<string, string>): Promise<FileResponse> {
    return new Promise((resolve, reject) => {
      const request = new UpdateFileMetadataRequest();
      request.setKey(fileKey);
      request.setMetadataMap(metadata);

      this.client.updateFileMetadata(request, this.getMetadata(), (error: grpc.Error, response: FileResponse) => {
        if (error) {
          reject(new Error(`Update file metadata failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  // Peer Operations

  async listPeers(): Promise<ListPeersResponse> {
    return new Promise((resolve, reject) => {
      this.client.listPeers({}, this.getMetadata(), (error: grpc.Error, response: ListPeersResponse) => {
        if (error) {
          reject(new Error(`List peers failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async getPeer(peerId: string): Promise<PeerResponse> {
    return new Promise((resolve, reject) => {
      const request = new PeerRequest();
      request.setId(peerId);

      this.client.getPeer(request, this.getMetadata(), (error: grpc.Error, response: PeerResponse) => {
        if (error) {
          reject(new Error(`Get peer failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async addPeer(address: string, port: number, metadata: Record<string, string> = {}): Promise<PeerResponse> {
    return new Promise((resolve, reject) => {
      const request = new AddPeerRequest();
      request.setAddress(address);
      request.setPort(port);
      request.setMetadataMap(metadata);

      this.client.addPeer(request, this.getMetadata(), (error: grpc.Error, response: PeerResponse) => {
        if (error) {
          reject(new Error(`Add peer failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async removePeer(peerId: string): Promise<RemovePeerResponse> {
    return new Promise((resolve, reject) => {
      const request = new PeerRequest();
      request.setId(peerId);

      this.client.removePeer(request, this.getMetadata(), (error: grpc.Error, response: RemovePeerResponse) => {
        if (error) {
          reject(new Error(`Remove peer failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async getPeerHealth(peerId: string): Promise<PeerHealthResponse> {
    return new Promise((resolve, reject) => {
      const request = new PeerRequest();
      request.setId(peerId);

      this.client.getPeerHealth(request, this.getMetadata(), (error: grpc.Error, response: PeerHealthResponse) => {
        if (error) {
          reject(new Error(`Get peer health failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  // System Operations

  async getSystemInfo(): Promise<SystemInfoResponse> {
    return new Promise((resolve, reject) => {
      this.client.getSystemInfo({}, this.getMetadata(), (error: grpc.Error, response: SystemInfoResponse) => {
        if (error) {
          reject(new Error(`Get system info failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async getMetrics(): Promise<MetricsResponse> {
    return new Promise((resolve, reject) => {
      this.client.getMetrics({}, this.getMetadata(), (error: grpc.Error, response: MetricsResponse) => {
        if (error) {
          reject(new Error(`Get metrics failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  async healthCheck(): Promise<HealthResponse> {
    return new Promise((resolve, reject) => {
      this.client.healthCheck({}, this.getMetadata(), (error: grpc.Error, response: HealthResponse) => {
        if (error) {
          reject(new Error(`Health check failed: ${error.message}`));
        } else {
          resolve(response);
        }
      });
    });
  }

  // Streaming Operations

  streamFileOperations(callback: (event: FileOperationEvent) => void): () => void {
    const stream = this.client.streamFileOperations({}, this.getMetadata());

    stream.on('data', (event: FileOperationEvent) => {
      callback(event);
    });

    stream.on('error', (error: grpc.Error) => {
      console.error('File operations stream error:', error);
    });

    return () => stream.cancel();
  }

  streamPeerEvents(callback: (event: PeerEvent) => void): () => void {
    const stream = this.client.streamPeerEvents({}, this.getMetadata());

    stream.on('data', (event: PeerEvent) => {
      callback(event);
    });

    stream.on('error', (error: grpc.Error) => {
      console.error('Peer events stream error:', error);
    });

    return () => stream.cancel();
  }

  streamSystemEvents(callback: (event: SystemEvent) => void): () => void {
    const stream = this.client.streamSystemEvents({}, this.getMetadata());

    stream.on('data', (event: SystemEvent) => {
      callback(event);
    });

    stream.on('error', (error: grpc.Error) => {
      console.error('System events stream error:', error);
    });

    return () => stream.cancel();
  }

  // Utility Methods

  private getMetadata(): grpc.Metadata {
    const metadata = new grpc.Metadata();
    metadata.append('authorization', `Bearer ${this.config.authToken}`);
    return metadata;
  }

  private calculateChecksum(data: Uint8Array): string {
    // Simple checksum calculation - in production, use a proper hash function
    let checksum = 0;
    for (let i = 0; i < data.length; i++) {
      checksum = (checksum + data[i]) % 256;
    }
    return checksum.toString(16);
  }
}

// Export types for convenience
export {
  FileRequest,
  FileResponse,
  ListFilesRequest,
  ListFilesResponse,
  DeleteFileResponse,
  UpdateFileMetadataRequest,
  AddPeerRequest,
  PeerRequest,
  PeerResponse,
  ListPeersResponse,
  RemovePeerResponse,
  PeerHealthResponse,
  SystemInfoResponse,
  MetricsResponse,
  HealthResponse,
  FileOperationEvent,
  PeerEvent,
  SystemEvent,
  FileChunk,
};
