# go-ini

Read key and Value pair from a ini file. This package keeps a watch on the file specified and 
notifies a call back function (if registered) for any value change.

APIs:

NewIniFile()  # Pass in the fileName to keep read key(s)/value, and keep a watch on.

ReadKey()     # Specify section for the file, key to be read and defaultval if no key is present.

Register()    # Register a cb function to be called when there is a change in the file being watched.
