# find-luna

This is a tool to find IBC UST/LUNA holders at specific heights on the Secret Network. 

At the time this was needed export in secretd wasn't working, so this uses grpc. If run locally on an archive node with a few extra lookup workers it should finish in less than 5 minutes (which is actually a lot faster than exporting state.)

The .csv files in this directory are the output results. 