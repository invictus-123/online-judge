package com.online.judge.executorv2.queue;

import static com.online.judge.executorv2.config.RabbitMqConfig.SUBMISSIONS_QUEUE;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.executorv2.model.message.SubmissionMessage;
import com.rabbitmq.client.Channel;
import java.io.IOException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.support.AmqpHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.stereotype.Component;

@Component
public class SubmissionListener {
	private static final Logger log = LoggerFactory.getLogger(SubmissionListener.class);

	private final ObjectMapper objectMapper;

	public SubmissionListener(ObjectMapper objectMapper) {
		this.objectMapper = objectMapper;
	}

	@RabbitListener(queues = SUBMISSIONS_QUEUE)
	public void handleSubmission(String message, Channel channel, @Header(AmqpHeaders.DELIVERY_TAG) long tag)
			throws IOException {
		try {
			SubmissionMessage submissionMessage = objectMapper.readValue(message, SubmissionMessage.class);
			log.info("Received submission {}", submissionMessage.submissionId());

			// TODO: Implement submission processing logic
			processSubmission(submissionMessage);

			channel.basicAck(tag, false);
			log.debug("ACK sent for submission {}", submissionMessage.submissionId());
		} catch (Exception e) {
			log.error("Error processing submission: {}", e.getMessage());
			channel.basicNack(tag, false, false);
			log.warn("NACK sent for submission");
		}
	}

	private void processSubmission(SubmissionMessage submissionMessage) {
		log.info("Processing submission {} with language {} and {} test cases",
			submissionMessage.submissionId(),
			submissionMessage.language(),
			submissionMessage.testCases().size());
	}
}