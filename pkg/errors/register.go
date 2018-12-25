package errors

func Register(errMap map[FGErrorCode]string) error {
	var ok bool
	if errMap == nil {
		return nil
	}
	for k, v := range errMap {
		if _, ok = evtDesc[k]; ok {
			return FGEOrderNameExist
		}
		evtDesc[k] = v
	}
	return nil
}
