 ```mermaid
sequenceDiagram
    participant P as Proposer
    participant A as Node A
    participant B as Node B
    participant C as Node C

    P->>P: Split data into N shards
    P->>P: Build Merkle tree
    P->>P: Generate ProofRequest for each shard

    P->>A: ProofRequest(shard proof)
    P->>B: ProofRequest(shard proof)
    P->>C: ProofRequest(shard proof)

    A->>A: Verify proof
    B->>B: Verify proof
    C->>C: Verify proof

    A-->>P: EchoRequest
    A-->>B: EchoRequest
    A-->>C: EchoRequest

    B-->>P: EchoRequest
    B-->>A: EchoRequest
    B-->>C: EchoRequest

    C-->>P: EchoRequest
    C-->>A: EchoRequest
    C-->>B: EchoRequest

    A->>A: If Echo >= N-F, broadcast Ready
    B->>B: If Echo >= N-F, broadcast Ready
    C->>C: If Echo >= N-F, broadcast Ready

    A-->>P: ReadyRequest
    A-->>B: ReadyRequest
    A-->>C: ReadyRequest

    B-->>P: ReadyRequest
    B-->>A: ReadyRequest
    B-->>C: ReadyRequest

    C-->>P: ReadyRequest
    C-->>A: ReadyRequest
    C-->>B: ReadyRequest

    A->>A: If Ready >= 2F+1 and Echo >= F+1, reconstruct
    B->>B: If Ready >= 2F+1 and Echo >= F+1, reconstruct
    C->>C: If Ready >= 2F+1 and Echo >= F+1, reconstruct
```