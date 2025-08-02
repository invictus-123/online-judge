package com.online.judge.backend.service;

import com.online.judge.backend.dto.message.ResultNotificationMessage;
import com.online.judge.backend.repository.TestCaseResultRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** Service class for managing test case results. */
@Service
public class TestCaseResultService {
	private static final Logger logger = LoggerFactory.getLogger(TestCaseResultService.class);

	private TestCaseResultRepository testCaseResultRepository;

	public TestCaseResultService(TestCaseResultRepository testCaseResultRepository) {
		this.testCaseResultRepository = testCaseResultRepository;
	}

	/**
	 * Updates the status for a particular submission.
	 *
	 * @param submissionId
	 *             The ID of the submission
	 * @param staus
	 *             The status of the submission
	 */
	@Transactional
	public void processTestResult(ResultNotificationMessage message) {}
}
