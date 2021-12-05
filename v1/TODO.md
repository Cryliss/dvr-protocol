# TODO

## /app
- [x] Need to add comments to parseInput

## /server
- [x] parse topology file.
- [x] update: update the link cost between two *neighboring* servers.
    - [x] auto updates (`s.Loopy`)
    - [x] command update
- [x] step: send routing update to neighbors right away.
- [x] packets: display the number of packets this server has received since this function was last called.
- [x] display: displays the current routing table
- [x] disable: disables the link to the given server, if it is its neighbor
- [x] crash: "close" all connections. meant to simulate server crashes. Neighboring servers must handle this close correctly and set the link cost to infinity

Fix disabling links that haven't sent an update in 3 update intervals ...

# Messages ✅
marshaling and unmarshaling the IP addresses needs to be redone ...
Current way  
1. breaks if the length of the IP address changes
2. marshals it into 12 bytes .. 3x larger than what it should be.

I believe I can try to split the address up, convert each address part into uint8's
Marshal each part

Then unmarshal each as uint8  
Then convert those into strings and combine into one string with "." separating each  

Cause the IP address should have four parts  
And each value as an integer should fit into a byte and we're allowed 4 bytes ..  



# Update command
When I use the command `update 3 4 1` on server 1, the change does not remain.  
But, when I use `update 3 4 1` on server 3 or 4, the change does remain & the link cost to server 2 is updated accordingly ..

is that how it should work?


# Disable

When disabling a link, should it be able to be re-enabled, if not all the servers have it as disabled?
.




s.t.mu.Lock()
rt := s.t.Routing
nb, ok := s.t.Neighbors[int(senderID)]
s.t.mu.Unlock()

// Loop through each of our message neighbors and update the routing table
// for the neighbor and the neighbor link costs accordingly.
for _, n := range msg.n {
    rt[x][int(n.nID)-1] = int(n.nCost)
    if ok {
        nb.mu.Lock()
        nb.Cost = int(n.nCost)
        if !nb.disabled {
            rt[int(n.nID)-1][x] = int(n.nCost)
        }
        nb.mu.Unlock()
    }
}

if ok {
    nb.mu.Lock()
    if nb.disabled {
        nb.disabled = false
    }
    nb.ts = time.Now()
    nb.mu.Unlock()
}
