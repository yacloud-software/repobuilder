
there is no means to 'reset' a git repository atm.
it's safe like that but yet...

to do so:

select * from createwebrequest;
NOTE THE ID

delete from repocreatestatus where createrequestid = [ID];

update trackergitrepository set sourceinstalled=false,protosubmitted=false,protocommitted=false,finalised=false,patchrepo=false,repositorycreated=false where createrequestid = [ID];



