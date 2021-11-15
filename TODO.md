# TODO

## /app
- [x] Need to add comments to parseInput

## /server
- [x] parse topology file.
- [ ] update: update the link cost between two *neighboring* servers.
- [ ] step: send routing update to neighbors right away.
- [x] packets: display the number of packets this server has received since this function was last called.
- [x] display: displays the current routing table
- [x] disable: disables the link to the given server, if it is its neighbor
- [ ] crash: "close" all connections. meant to simulate server crashes. Neighboring servers must handle this close correctly and set the link cost to infinity



# Messages
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
