package main

type Session struct {
	seq     int32
	servers map[int32]int32 // seq:server
	// room:seqs
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

func (s *Session) Servers() (seqs []int32, servers []int32) {
	var (
		i           = len(s.servers)
		seq, server int32
	)
	seqs = make([]int32, i)
	servers = make([]int32, i)
	for seq, server = range s.servers {
		i--
		seqs[i] = seq
		servers[i] = server
	}
	return
}

// Del delete the session by sub key.
func (s *Session) Del(seq int32) bool {
	delete(s.servers, seq)
	return (len(s.servers) == 0)
}

func (s *Session) Size() int {
	return len(s.servers)
}
