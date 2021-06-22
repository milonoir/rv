package common

// Widget provides an interface to interact with widgets.
// A widget renders some sort of data to a termui widget.
type Widget interface {
	// Update is called by the application periodically and implementations must render their data
	// to a termui widget.
	Update()

	// Close is called when the application exists. Widgets can implement this to return gracefully.
	Close()
}