type Relay interface {
        Listen
        Connect
        UpwardRead
        UpwardWrite
        DownwardRead
        DownwardWrite
}

type SSLocal struct {
}

type SSServer struct {
}

