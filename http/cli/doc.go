// Provides a standard interface to start a net/http.Handler with a configurable http server.
// When registering with a FlagSet (compatible to package flag or github.com/ogier/pflag) you can configure
// the server via the commandline.
//
// Example:
//    starter := httpcli.NewStarterFromFlagSet(flag.CommandLine)
//    flag.Parse()
//    starter.StartHttpInterface(myHandler)
//
package cli
