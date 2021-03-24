package util

// 顺序执行，遇到err退出
// 并发执行使用
func TryError(steps ...func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
