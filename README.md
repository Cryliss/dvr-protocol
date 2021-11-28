# dvr-protocol
 Repository for COMP 429 Programming Assignment#2 - Distance Vector Routing Protocols


# Assignment Details
In this assignment you will implement a simplified version of the *Distance Vector Routing Protocol*.  
The protocol will be run on top of **four** servers/laptops (behaving as routers) using **TCP** or **UDP**.  
Each server runs on a machine at a pre-defined port number.  
The servers should be able to output their forwarding tables along with the cost and should be robust to link changes.  
A Server should send out routing packets only in the following two conditions:  
    a) periodic update  
    b) the user uses a command asking for one  
*This is a little different from the original algorithm which immediately sends out update routing information when the routing table changes.*


# Protocol Specification
The various components of the protocol are explained step by step. Please strictly adhere to the specifications.

## Topology Establishment
The four servers are required to form a network topology as shown in Fig. 1.
![Figure 1: Example Topology](https://github.com/Cryliss/dvr-protocol/docs/Figure-1-Example-Topology.png)

Each server is supplied with a topology file at startup that it uses to build its initial routing table.  
The topology file is local and contains the link cost to the neighbors (all other servers will be infinity).  
Each server can only read the topology file for itself.

The entries of a topology file are listed below:

- `<num-servers>`
- `<num-neighbors>`
- `<server-ID> <server-IP> <server-port>`
- `<server-ID1> <server-ID2> <cost>`

**num-servers**: total number of servers.  
**server-ID**, **server-ID1**, **server-ID2**: a unique identifier for a server, which is assigned by you.  
**cost**: cost of a given link between a pair of servers. Assume that cost is an integer value.  

For cost values, each topology file should only contain the cost values of the host server’s neighbors.  

### IMPORTANT
In this environment, costs are bi-directional i.e. the cost of a link from A-B is the same for B-A.  
Whenever a new server is added to the network, it will read its topology file to determine who are its neighbors.  

Routing updates are exchanged periodically between neighboring servers.  
When this newly added server sends routing messages to its neighbors, they will add an entry in their routing tables corresponding to it. Servers can also be removed from a network.  

When a server has been removed from a network, it will no longer send distance vector updates to its neighbors.

When a server *no longer receives distance vector updates from its neighbor for three consecutive update intervals*, it **assumes that the neighbor no longer exists** in the network and makes the appropriate changes to its routing table (link cost to this neighbor will now be set to infinity but not remove it from the table).

This information is propagated to other servers in the network with the exchange of routing updates.  

##  Routing Updates
Routing updates are exchanged periodically between neighboring servers based on a time interval specified at the startup.   In addition to exchanging distance vector updates, servers must also be able to respond to user-specified events.  

There are 3 possible events in this system.
They can be grouped into three classes:
(1) **Topology** changes refer to an updating of link status (**update**).  
(2) **Queries** include the ability to ask a server for its current routing table (**display**), and to ask a server for the number of distance vectors it has received (**packets**). In the case of the packets command, the value is reset to **zero by** a server after it satisfies the query.  
(3) **Exchange commands** can cause a server to send distance vectors to its neighbors immediately.

## Message Format

Routing updates are sent using the General Message format. All routing updates are **UDP unreliable messages**.  

The message format for the data part is:  

**Number of update fields**: (2 bytes):Indicate the number of entries that follow.
**Server port**: (2 bytes) port of the server sending this packet.
**Server IP**: (4 bytes) IP of the server sending this packet.
**Server IP address n**: (4 bytes) IP of the n-th server in its routing table.
**Server port n**: (2 bytes) port of the n-th server in its routing table.
**Server IDn**: (2 bytes) server id of the n-th server on the network.
**Cost n**: cost of the **path** from the server sending the update to the n-th server whose ID is given in the packet.

### Note
First, the servers listed in the packet can be any order i.e., 5,3, 2, 1, 4.  
Second, the packet needs to include an entry to reach itself with cost 0  
    i.e. server 1 needs to have an entry of cost 0 to reach server 1.

# Server Commands / Input Format

## Startup
The server must support the following command at startup:
`server -t <topology-file-name> -i <routing-update-interval>`

**topology-file-name**: The topology file contains the initial topology configuration for the server, e.g., timberlake_init.txt.  
**routing-update-interval**: It specifies the time interval between routing updates in seconds.  
**port and server-id**: They are written in the topology file. The server should find its port and server-id in the topology file without changing the entry format or adding any new entries.

## Run Time
The following commands can be specified at any point during the run of the server:

1. `update <server-ID1> <server-ID2> <Link Cost>`  
**server-ID1, server-ID2**: The link for which the cost is being updated.  
**Link Cost**: It specifies the new link cost between the source and the destination server. Note that this command will be issued to **both** *server-ID1* and *server-ID2* and involve them to update the cost and no other server.  

For example:  
- `update 1 2 inf`: The link between the servers with IDs 1 and 2 is assigned to infinity.   
- `update 1 2 8`: Change the cost of the link to 8.  

2. `step`  
Send routing update to neighbors right away.

3. `packets`  
Display the number of packets this server has received since the last invocation of this command.

4. `display`  
Display the current routing table

The table should be displayed in a **sorted** order from small ID to big ID.  
The display should be formatted as a sequence of lines, with each line indicating: `<source-server-ID> <next-hop-server-ID> <cost-of-path>`  

5. `disable<server-ID>`    
Disable the link to a given server. Doing this “closes” the connection to a given server with server-ID.  
Here you need to check if the given server is its neighbor.  

6. `crash`  
“Close” all connections on all links. This is to simulate server crashes.  
The neighboring servers must handle this close correctly and set the link cost to infinity.

# Server Responses / Output Format
The following are a list of possible responses a user can receive from a server:

1. On successful execution of an update, step, packets, display or disable command, the server must display the following message:
    `<command-string> SUCCESS`  

where command-string is the command executed.

2. Upon encountering an error during execution of one of these commands, the server must display the following response:  
    `<command-string> <error message> `
where error message is a brief description of the error encountered.

3. On successfully receiving a route update message from neighbors, the server must display the following response:  

    `RECEIVED A MESSAGE FROM SERVER <server-ID>`  
Where the server-ID is the id of the server which sent a route update message to the local server.
