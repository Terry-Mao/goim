package main

type Session struct {
	seq     int32
	servers map[int32]int32 // map[user_id] ->  map[sub_id] -> server_id
}

// NewSession new a session struct. store the seq and serverid.
func NewSession(server int) *Session {
	s := new(Session)
	s.servers = make(map[int32]int32, server)
	s.seq = 0
	return s
}

func (s *Session) nextSeq() int32 {
	s.seq++
	return s.seq
}

// Put put a session according with sub key.
func (s *Session) Put(server int32) (seq int32) {
	seq = s.nextSeq()
	s.servers[seq] = server
	return
}

func (s *Session) Servers() map[int32]int32 {
	// must readonly
	return s.servers
}

// Del delete the session by sub key.
func (s *Session) Del(seq int32) bool {
	delete(s.servers, seq)
	return (len(s.servers) == 0)
}

func (s *Session) Size() int {
	return len(s.servers)
}
