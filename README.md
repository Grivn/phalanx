# Phalanx

## Introduction
Generally, a BFT (*Byzantine Fault Tolerance*) consensus algorithm usually concentrates on two properties, *safety* and *liveness*. 
Well, in the system on reality, we should also pay attention to the order of transactions, that is an incorrect order may make the task failed. 
However, the BFT algorithm, especially the partial-synchronized protocols, cannot deal with such a problem because of the leader-based model.
The participates could find out the malicious manners of leader about the sequence order, such as the 
fork-attack or silence-attack, but the leader could always make a decision on the content of proposals, 
and it could decide the order the transactions on purpose to make them failed, and such an attack manner cannot be detected.
So that, in *CRYPTO2020*, Miller has introduced a new kind of property called *order-fairness*, which indicates
a protocol's property to make a trusted order for transactions. 

*Phalanx* is an order-fairness byzantine fault tolerance protocol. 
It could complete order-fairness property, and become a plugin for most kinds of BFT protocol, 
which means a traditional BFT protocol could complete the order-fairness property easily with the accession of *Phalanx*.
In addition, the Phalanx protocol could also be used in asynchronized BFT protocols, such as HoneyBadger BFT or Dumbo,
to complete the transaction order in blocks, which is an open problem for them.

## Background
The development of BFT protocols, the importance for order-fairness property, and some research on it.

## Phalanx Phase (need update)
### Overview

### Steps
#### Proposal-Generation
Every node for phalanx protocol could generate proposals.
When a node has found some transactions of batch-size or reached batch timer interval, 
it would like to generate a proposal and assign a sequence number for it.
- <PROPOSAL n, block, id, sig>

#### Sequence-Assignment
Every node would like assign a specific sequence number for the proposal it has received, 
and generate a log message to notify others.
- <LOG n, block-identifier, id, sig>

#### Aggregate-Certification
A node would like to start a specific voter for every participates in phalanx cluster.
- <VOTE log-identifier, id, sig>

#### Byzantine Consensus Component
Nodes in phalanx cluster would like to make a consensus on the selected qc on specific sequence number.

#### Execution
When node has found there are quorum participates assigned one proposal, it could be selected in next block.
If there are more than 2 proposals selected in one block, we would like to order them according to timestamp info.

### Proof
#### Safety
Proof according to reliable broadcast.

#### Liveness
Proof according to reliable broadcast.

#### Order-Fairness