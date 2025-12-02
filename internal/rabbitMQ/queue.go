package rabbitmq

func SetupQueue() error {
	queues := []string{"email", "sms"}

	for _, q := range queues {
		_, err := Ch.QueueDeclare(q+"_queue", true, false, false, false, nil)
		if err != nil {
			return err
		}

		_, err_dlq := Ch.QueueDeclare(q+"_dlq", true, false, false, false, nil)
		if err_dlq != nil {
			return err_dlq
		}
	}

	return nil
}
