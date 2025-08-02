package com.online.judge.backend.config;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.DirectExchange;
import org.springframework.amqp.core.FanoutExchange;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.core.QueueBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitMqConfig {
	public static final String SUBMISSIONS_EXCHANGE = "oj.ex.submissions";
	public static final String STATUS_EXCHANGE = "oj.ex.status";
	public static final String RESULTS_EXCHANGE = "oj.ex.results";

	public static final String SUBMISSIONS_DLX = "oj.ex.submissions.dlx";

	public static final String SUBMISSIONS_QUEUE = "oj.q.submissions";
	public static final String STATUS_QUEUE = "oj.q.status";
	public static final String RESULTS_QUEUE = "oj.q.results";

	public static final String SUBMISSIONS_RETRY_QUEUE = "oj.q.submissions.retry";
	public static final String SUBMISSIONS_FAILED_QUEUE = "oj.q.submissions.failed";

	public static final String SUBMISSION_NEW_KEY = "submission.new";
	public static final String SUBMISSION_STATUS_KEY = "submission.status";
	public static final String SUBMISSION_RESULT_KEY = "submission.result";

	@Value("${rabbitmq.retry.ttl:30000}")
	private Integer retryTtl;

	@Bean
	DirectExchange submissionsExchange() {
		return new DirectExchange(SUBMISSIONS_EXCHANGE);
	}

	@Bean
	DirectExchange statusExchange() {
		return new DirectExchange(STATUS_EXCHANGE);
	}

	@Bean
	DirectExchange resultsExchange() {
		return new DirectExchange(RESULTS_EXCHANGE);
	}

	@Bean
	Queue submissionsQueue() {
		return QueueBuilder.durable(SUBMISSIONS_QUEUE)
				.withArgument("x-dead-letter-exchange", SUBMISSIONS_DLX)
				.build();
	}

	@Bean
	Binding submissionsBinding(Queue submissionsQueue, DirectExchange submissionsExchange) {
		return BindingBuilder.bind(submissionsQueue).to(submissionsExchange).with(SUBMISSION_NEW_KEY);
	}

	@Bean
	Queue statusQueue() {
		return new Queue(STATUS_QUEUE, true);
	}

	@Bean
	Binding statusBinding(Queue statusQueue, DirectExchange statusExchange) {
		return BindingBuilder.bind(statusQueue).to(statusExchange).with(SUBMISSION_STATUS_KEY);
	}

	@Bean
	Queue resultsQueue() {
		return new Queue(RESULTS_QUEUE, true);
	}

	@Bean
	Binding resultsBinding(Queue resultsQueue, DirectExchange resultsExchange) {
		return BindingBuilder.bind(resultsQueue).to(resultsExchange).with(SUBMISSION_RESULT_KEY);
	}

	@Bean
	FanoutExchange submissionsDlx() {
		return new FanoutExchange(SUBMISSIONS_DLX);
	}

	@Bean
	Queue submissionsRetryQueue() {
		return QueueBuilder.durable(SUBMISSIONS_RETRY_QUEUE)
				.withArgument("x-dead-letter-exchange", SUBMISSIONS_EXCHANGE)
				.withArgument("x-dead-letter-routing-key", SUBMISSION_NEW_KEY)
				.withArgument("x-message-ttl", retryTtl)
				.build();
	}

	@Bean
	Binding submissionsRetryBinding(Queue submissionsRetryQueue, FanoutExchange submissionsDlx) {
		return BindingBuilder.bind(submissionsRetryQueue).to(submissionsDlx);
	}

	@Bean
	Queue submissionsFailedQueue() {
		return new Queue(SUBMISSIONS_FAILED_QUEUE, true);
	}
}
