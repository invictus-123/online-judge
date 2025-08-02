package com.online.judge.backend.queue;

import static com.online.judge.backend.config.RabbitMqConfig.STATUS_QUEUE;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.backend.dto.message.StatusUpdateMessage;
import com.online.judge.backend.service.SubmissionService;
import com.rabbitmq.client.Channel;
import java.io.IOException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.support.AmqpHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.stereotype.Component;

@Component
public class StatusListener {
	private static final Logger log = LoggerFactory.getLogger(StatusListener.class);

	private final ObjectMapper objectMapper;
	private final SubmissionService submissionService;

	public StatusListener(ObjectMapper objectMapper, SubmissionService submissionService) {
		this.objectMapper = objectMapper;
		this.submissionService = submissionService;
	}

	@RabbitListener(queues = STATUS_QUEUE)
	public void handleStatusUpdate(String message, Channel channel, @Header(AmqpHeaders.DELIVERY_TAG) long tag)
			throws IOException {
		StatusUpdateMessage statusUpdateMessage = objectMapper.readValue(message, StatusUpdateMessage.class);
		log.info(
				"Received status update for submission {}: {}",
				statusUpdateMessage.submissionId(),
				statusUpdateMessage.status());
		try {
			submissionService.updateStatus(statusUpdateMessage.submissionId(), statusUpdateMessage.status());

			channel.basicAck(tag, /* multiple= */ false);
			log.debug("ACK sent for status update of submission {}", statusUpdateMessage.submissionId());
		} catch (Exception e) {
			log.error(
					"Error processing status update for submission {}: {}",
					statusUpdateMessage.submissionId(),
					e.getMessage());

			channel.basicNack(tag, false, false);
			log.warn("NACK sent for status update of submission {}", statusUpdateMessage.submissionId());
		}
	}
}
