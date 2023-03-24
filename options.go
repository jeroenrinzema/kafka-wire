package wire

// OptionFn options pattern used to define and set options for the given
// Kafka server.
type OptionFn func(*Server) error
