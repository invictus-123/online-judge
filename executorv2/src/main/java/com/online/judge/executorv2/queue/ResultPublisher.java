package com.online.judge.executorv2.queue;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.executorv2.config.RabbitMqConfig;
import com.online.judge.executorv2.model.message.ExecutionResultMessage;
import com.online.judge.executorv2.model.message.StatusUpdateMessage;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.connection.CorrelationData;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.lang.Nullable;
import org.springframework.stereotype.Service;

@Service
public class ResultPublisher implements RabbitTemplate.ConfirmCallback {
	private static final Logger log = LoggerFactory.getLogger(ResultPublisher.class);

	private final RabbitTemplate rabbitTemplate;
	private final ObjectMapper objectMapper;

	public ResultPublisher(RabbitTemplate rabbitTemplate, ObjectMapper objectMapper) {
		this.rabbitTemplate = rabbitTemplate;
		this.objectMapper = objectMapper;
	}

	@PostConstruct
	public void init() {
		rabbitTemplate.setConfirmCallback(this);
	}

	public void sendStatusUpdate(StatusUpdateMessage statusMessage) {
		try {
			CorrelationData correlationData = new CorrelationData(String.valueOf(statusMessage.submissionId()));

			rabbitTemplate.convertAndSend(
					RabbitMqConfig.STATUS_EXCHANGE,
					RabbitMqConfig.SUBMISSION_STATUS_KEY,
					objectMapper.writeValueAsString(statusMessage),
					correlationData);
			log.info("Status update for submission {} published to RabbitMQ.", statusMessage.submissionId());
		} catch (Exception e) {
			log.error("Failed to publish status update for submission {}: {}", statusMessage.submissionId(), e.getMessage());
		}
	}

	public void sendExecutionResult(ExecutionResultMessage resultMessage) {
		try {
			CorrelationData correlationData = new CorrelationData(String.valueOf(resultMessage.submissionId()));

			rabbitTemplate.convertAndSend(
					RabbitMqConfig.RESULTS_EXCHANGE,
					RabbitMqConfig.SUBMISSION_RESULT_KEY,
					objectMapper.writeValueAsString(resultMessage),
					correlationData);
			log.info("Execution result for submission {} published to RabbitMQ.", resultMessage.submissionId());
		} catch (Exception e) {
			log.error("Failed to publish execution result for submission {}: {}", resultMessage.submissionId(), e.getMessage());
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
		}
	}
}