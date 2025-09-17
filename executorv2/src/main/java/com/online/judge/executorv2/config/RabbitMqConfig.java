package com.online.judge.executorv2.config;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.DirectExchange;
import org.springframework.amqp.core.FanoutExchange;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.core.QueueBuilder;
import org.springframework.amqp.rabbit.config.SimpleRabbitListenerContainerFactory;
import org.springframework.amqp.rabbit.connection.ConnectionFactory;
import org.springframework.amqp.support.converter.Jackson2JsonMessageConverter;
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

	public static final String SUBMISSION_NEW_KEY = "submission.new";
	public static final String SUBMISSION_STATUS_KEY = "submission.status";
	public static final String SUBMISSION_RESULT_KEY = "submission.result";

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
	Queue statusQueue() {
		return new Queue(STATUS_QUEUE, true);
	}

	@Bean
	Queue resultsQueue() {
		return new Queue(RESULTS_QUEUE, true);
	}

	@Bean
	Binding submissionsBinding(Queue submissionsQueue, DirectExchange submissionsExchange) {
		return BindingBuilder.bind(submissionsQueue).to(submissionsExchange).with(SUBMISSION_NEW_KEY);
	}

	@Bean
	Binding statusBinding(Queue statusQueue, DirectExchange statusExchange) {
		return BindingBuilder.bind(statusQueue).to(statusExchange).with(SUBMISSION_STATUS_KEY);
	}

	@Bean
	Binding resultsBinding(Queue resultsQueue, DirectExchange resultsExchange) {
		return BindingBuilder.bind(resultsQueue).to(resultsExchange).with(SUBMISSION_RESULT_KEY);
	}

	@Bean
	Jackson2JsonMessageConverter messageConverter() {
		return new Jackson2JsonMessageConverter();
	}

	@Bean
	FanoutExchange submissionsDlx() {
		return new FanoutExchange(SUBMISSIONS_DLX);
	}

	@Bean
	SimpleRabbitListenerContainerFactory rabbitListenerContainerFactory(ConnectionFactory connectionFactory) {
		SimpleRabbitListenerContainerFactory factory = new SimpleRabbitListenerContainerFactory();
		factory.setConnectionFactory(connectionFactory);
		factory.setMessageConverter(messageConverter());
		factory.setAcknowledgeMode(org.springframework.amqp.core.AcknowledgeMode.MANUAL);
		return factory;
	}
}