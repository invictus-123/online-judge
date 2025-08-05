package com.online.judge.backend.controller;

import com.online.judge.backend.annotation.RateLimit;
import com.online.judge.backend.dto.filter.SubmissionFilterRequest;
import com.online.judge.backend.dto.request.SubmitCodeRequest;
import com.online.judge.backend.dto.response.GetSubmissionByIdResponse;
import com.online.judge.backend.dto.response.ListSubmissionsResponse;
import com.online.judge.backend.dto.response.SubmitCodeResponse;
import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.service.SubmissionService;
import jakarta.validation.Valid;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

/** REST controller for handling submission-related requests. */
@RestController
@RequestMapping("/api/v1/submissions")
public class SubmissionController {
	private static final Logger logger = LoggerFactory.getLogger(SubmissionController.class);

	private final SubmissionService submissionService;

	public SubmissionController(SubmissionService submissionService) {
		this.submissionService = submissionService;
	}

	/**
	 * Handles GET requests to fetch a list of submissions with optional filtering and pagination,
	 * sorted by submission date in descending order.
	 *
	 * @param page The page number to retrieve (default is 1).
	 * @param onlyMe Filter to show only submissions by the current user (optional).
	 * @param problemId Filter submissions by problem ID (optional).
	 * @param statuses Filter submissions by status (optional).
	 * @param languages Filter submissions by programming language (optional).
	 * @return A ResponseEntity containing a paginated list of submissions matching the filters.
	 */
	@GetMapping("/list")
	public ResponseEntity<ListSubmissionsResponse> listSubmissions(
			@RequestParam(defaultValue = "1") int page,
			@RequestParam(required = false) Boolean onlyMe,
			@RequestParam(required = false) Long problemId,
			@RequestParam(required = false, name = "status") List<SubmissionStatus> statuses,
			@RequestParam(required = false, name = "language") List<SubmissionLanguage> languages) {

		SubmissionFilterRequest filterRequest =
				new SubmissionFilterRequest(onlyMe, problemId, statuses, languages, page);
		logger.info("Received call to fetch submissions with filters={}", filterRequest);

		ListSubmissionsResponse response =
				new ListSubmissionsResponse(submissionService.listSubmissions(filterRequest));
		return ResponseEntity.ok(response);
	}

	/**
	 * Handles GET requests to fetch the details of a specific submission by its ID.
	 *
	 * @param id
	 *          The ID of the submission to retrieve.
	 * @return A ResponseEntity containing the SubmissionDetailsUi DTO if found, or a 404 Not Found
	 *         status if the submission does not exist.
	 */
	@GetMapping("/{id}")
	public ResponseEntity<GetSubmissionByIdResponse> getSubmissionById(@PathVariable Long id) {
		logger.info("Received call to fetch submission with ID {}", id);

		GetSubmissionByIdResponse response =
				new GetSubmissionByIdResponse(submissionService.getSubmissionDetailsById(id));
		return ResponseEntity.ok(response);
	}

	/**
	 * Handles POST requests to make a submission. Only accessible by authenticated users.
	 *
	 * @param request
	 *               The request body containing the submission details.
	 * @return A ResponseEntity with the created submission and a 201 Created status. Throws a 401
	 *         error if the user is not authorized to make a submission.
	 */
	@PostMapping
	@RateLimit(apiType = "submit-code", capacity = 10, refillPeriodMinutes = 1)
	public ResponseEntity<SubmitCodeResponse> submit(@Valid @RequestBody SubmitCodeRequest request) {
		logger.info("Received request to create a new submission");

		SubmitCodeResponse response = new SubmitCodeResponse(submissionService.submitCode(request));
		return new ResponseEntity<>(response, HttpStatus.CREATED);
	}
}
