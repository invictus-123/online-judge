package com.online.judge.backend.queue;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.backend.config.RabbitMqConfig;
import com.online.judge.backend.dto.message.SubmissionMessage;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.connection.CorrelationData;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.lang.Nullable;
import org.springframework.stereotype.Service;

@Service
public class SubmissionPublisher implements RabbitTemplate.ConfirmCallback {
	private static final Logger log = LoggerFactory.getLogger(SubmissionPublisher.class);

	private final RabbitTemplate rabbitTemplate;
	private final ObjectMapper objectMapper;

	public SubmissionPublisher(RabbitTemplate rabbitTemplate, ObjectMapper objectMapper) {
		this.rabbitTemplate = rabbitTemplate;
		this.objectMapper = objectMapper;
	}

	@PostConstruct
	public void init() {
		rabbitTemplate.setConfirmCallback(this);
	}

	public void sendSubmission(SubmissionMessage submissionMessage) {
		try {
			CorrelationData correlationData = new CorrelationData(String.valueOf(submissionMessage.submissionId()));

			rabbitTemplate.convertAndSend(
					RabbitMqConfig.SUBMISSIONS_EXCHANGE,
					RabbitMqConfig.SUBMISSION_NEW_KEY,
					objectMapper.writeValueAsString(submissionMessage),
					correlationData);
			log.info("Submission {} published to RabbitMQ.", submissionMessage.submissionId());
		} catch (Exception e) {
			log.error("Failed to publish submission {}: {}", submissionMessage.submissionId(), e.getMessage());
			// Handle publishing failure
		}
	}

	@Override
	public void confirm(@Nullable CorrelationData correlationData, boolean ack, @Nullable String cause) {
		if (correlationData == null) {
			return;
		}

		if (ack) {
			log.info("Publisher confirm ACK received for submission ID: {}", correlationData.getId());
		} else {
			log.error(
					"Publisher confirm NACK received for submission ID: {}. Cause: {}", correlationData.getId(), cause);
			// Implement retry logic or mark submission as failed in the database
		}
	}
}
