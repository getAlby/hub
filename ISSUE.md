Include fees in the balance that left a wallet listed on transaction list #464
Open
#1525
@stackingsaunter
Description
stackingsaunter
opened on Aug 13, 2024
Contributor

Currently, they are only visible in details.
Activity
stackingsaunter
added
good first issueGood for newcomers
on Aug 13, 2024
bumi
bumi commented on Aug 13, 2024
bumi
on Aug 13, 2024
Contributor

do you really want that? more numbers? what will it tell people and which other tranasction lists do that?
stackingsaunter
stackingsaunter commented on Aug 13, 2024
stackingsaunter
on Aug 13, 2024
ContributorAuthor

How is it more numbers? It's the same amount of numbers

Wallet that do that: Phoenix, Bitkey, Strike, Bitkit, Blink, Aqua

Wallets that don't do that: Muun, Blockstream Green, Umbrel
rolznz
rolznz commented on Aug 13, 2024
rolznz
on Aug 13, 2024
Contributor

The title of this issue is weird, but this is what you mean, right?

Use the sum of the amount + fee for the value shown on entries in the transaction list

Then when you open the details you would see the amount + fee separated
rolznz
rolznz commented on Aug 13, 2024
rolznz
on Aug 13, 2024
Contributor

@reneaaron brought up a good point - if we make this change it should be consistent across all Alby products (currently none are done this way)
GBKS
GBKS commented on Sep 16, 2024
GBKS
on Sep 16, 2024

I agree with @rolznz. Show the total amount that left your wallet in the transaction list, and show the breakdown in the details (maybe also in a hover tool tip on desktop).
stackingsaunter
stackingsaunter commented on Sep 16, 2024
stackingsaunter
on Sep 16, 2024
ContributorAuthor

Thanks @GBKS for your input, I also think this is best
stackingsaunter
stackingsaunter commented on Sep 16, 2024
stackingsaunter
on Sep 16, 2024
ContributorAuthor

In that case I would adopt this in all our products, and we could start with Hub? @rolznz
