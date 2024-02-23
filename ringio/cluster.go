package ringio

func (r *Ring) Join(addr string) error {
	return r.nodeService.Join(addr)
}
