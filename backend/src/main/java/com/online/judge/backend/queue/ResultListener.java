package com.online.judge.backend.queue;

import static com.online.judge.backend.config.RabbitMqConfig.RESULTS_QUEUE;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.backend.dto.message.ResultNotificationMessage;
import com.online.judge.backend.service.SubmissionService;
import com.online.judge.backend.service.TestCaseResultService;
import com.rabbitmq.client.Channel;
import java.io.IOException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.support.AmqpHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.stereotype.Component;

@Component
public class ResultListener {
	private static final Logger log = LoggerFactory.getLogger(ResultListener.class);

	private final ObjectMapper objectMapper;
	private final SubmissionService submissionService;
	private final TestCaseResultService testCaseResultService;

	public ResultListener(
			ObjectMapper objectMapper,
			SubmissionService submissionService,
			TestCaseResultService testCaseResultService) {
		this.objectMapper = objectMapper;
		this.submissionService = submissionService;
		this.testCaseResultService = testCaseResultService;
	}

	@RabbitListener(queues = RESULTS_QUEUE)
	public void handleFinalResult(String message, Channel channel, @Header(AmqpHeaders.DELIVERY_TAG) long tag)
			throws IOException {
		ResultNotificationMessage resultNotificationMessage =
				objectMapper.readValue(message, ResultNotificationMessage.class);
		log.info("Received final result for submission {}", resultNotificationMessage.submissionId());
		try {
			submissionService.updateTimeTakenAndMemoryUsed(
					resultNotificationMessage.submissionId(),
					resultNotificationMessage.timeTaken(),
					resultNotificationMessage.memoryUsed());
			testCaseResultService.processTestResult(
					resultNotificationMessage.submissionId(), resultNotificationMessage.testCaseResults());

			channel.basicAck(tag, /* multiple= */ false);
			log.debug("ACK sent for final result of submission {}", resultNotificationMessage.submissionId());
		} catch (Exception e) {
			log.error(
					"Error processing final result for submission {}: {}",
					resultNotificationMessage.submissionId(),
					e.getMessage());

			channel.basicNack(tag, /* multiple= */ false, /* requeue= */ false);
			log.warn("NACK sent for final result of submission {}", resultNotificationMessage.submissionId());
		}
	}
}
