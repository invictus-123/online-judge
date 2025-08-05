package com.online.judge.backend.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.request.CreateTestCaseRequest;
import com.online.judge.backend.dto.request.LoginRequest;
import com.online.judge.backend.dto.request.RegisterRequest;
import com.online.judge.backend.dto.request.SubmitCodeRequest;
import com.online.judge.backend.dto.response.AuthResponse;
import com.online.judge.backend.dto.response.CreateProblemResponse;
import com.online.judge.backend.dto.response.GetProblemByIdResponse;
import com.online.judge.backend.dto.response.GetSubmissionByIdResponse;
import com.online.judge.backend.dto.response.ListProblemsResponse;
import com.online.judge.backend.dto.response.ListSubmissionsResponse;
import com.online.judge.backend.dto.response.SubmitCodeResponse;
import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.dto.ui.SubmissionDetailsUi;
import com.online.judge.backend.dto.ui.SubmissionSummaryUi;
import com.online.judge.backend.dto.ui.UserSummaryUi;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.model.shared.UserRole;
import java.util.List;
import java.util.UUID;
import org.junit.jupiter.api.MethodOrderer;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestMethodOrder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

/**
 * Integration test for user and problem-related API flows. Uses Testcontainers to spin up a real
 * PostgreSQL database for testing. The tests are ordered because they represent a logical flow of
 * actions.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@Testcontainers
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
@Tag("integration")
@ActiveProfiles("integration")
class LifecycleIT {
	private static final String BASE_URL = "/api/v1";
	private static final String PROBLEM_TITLE = "Title";
	private static final String PROBLEM_STATEMENT = "Statement";
	private static final int PROBLEM_COUNT = 30;
	private static final ProblemDifficulty PROBLEM_DIFFICULTY = ProblemDifficulty.EASY;
	private static final Double PROBLEM_TIME_LIMIT = 1.0;
	private static final Integer PROBLEM_MEMORY_LIMIT = 256;
	private static final List<ProblemTag> PROBLEM_TAGS = List.of(ProblemTag.DP, ProblemTag.GRAPH);
	private static final List<CreateTestCaseRequest> PROBLEM_TEST_CASES = List.of(
			new CreateTestCaseRequest("Input 1", "Output 1", true, "Explanation"),
			new CreateTestCaseRequest("Input 2", "Output 2", false, null));
	private static final CreateProblemRequest CREATE_PROBLEM_REQUEST = new CreateProblemRequest(
			PROBLEM_TITLE,
			PROBLEM_STATEMENT,
			PROBLEM_DIFFICULTY,
			PROBLEM_TIME_LIMIT,
			PROBLEM_MEMORY_LIMIT,
			PROBLEM_TAGS,
			PROBLEM_TEST_CASES);
	private static final int SUBMISSION_COUNT = 25;
	private static final SubmissionLanguage SUBMISSION_LANGUAGE = SubmissionLanguage.CPP;
	private static final String CODE = "void main() {}";

	@Container
	private static final PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15-alpine");

	@DynamicPropertySource
	private static void configureProperties(DynamicPropertyRegistry registry) {
		registry.add("spring.datasource.url", postgres::getJdbcUrl);
		registry.add("spring.datasource.username", postgres::getUsername);
		registry.add("spring.datasource.password", postgres::getPassword);
		registry.add("spring.datasource.driver-class-name", () -> "org.postgresql.Driver");
		registry.add("spring.jpa.hibernate.ddl-auto", () -> "create");
	}

	@Autowired
	private TestRestTemplate restTemplate;

	private static String userHandle;
	private static String userToken;
	private static String adminHandle;
	private static String adminToken;
	private static Long problemId;
	private static Long submissionId;

	@Test
	@Order(1)
	void userAndAdminRegistrationAndLogin() {
		// 1. Register a standard USER
		userHandle = ("testuser-" + UUID.randomUUID()).substring(0, 20);
		String userEmail = "testuser-" + UUID.randomUUID() + "@example.com";
		String password = "password123";
		RegisterRequest userRegisterRequest =
				new RegisterRequest(userHandle, userEmail, password, "First", "Last", UserRole.USER);
		ResponseEntity<AuthResponse> userRegisterResponse =
				restTemplate.postForEntity(BASE_URL + "/auth/register", userRegisterRequest, AuthResponse.class);
		assertThat(userRegisterResponse.getStatusCode()).isEqualTo(HttpStatus.OK);

		// 2. Login as the registered USER and retrieve the JWT
		LoginRequest userLoginRequest = new LoginRequest(userHandle, password);
		ResponseEntity<AuthResponse> userLoginResponse =
				restTemplate.postForEntity(BASE_URL + "/auth/login", userLoginRequest, AuthResponse.class);
		assertThat(userLoginResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertThat(userLoginResponse.getBody()).isNotNull();
		userToken = userLoginResponse.getBody().token();
		assertThat(userToken).isNotBlank();

		// 3. Register an ADMIN and retrieve the JWT
		adminHandle = ("testadmin-" + UUID.randomUUID()).substring(0, 20);
		String adminEmail = "testadmin-" + UUID.randomUUID() + "@example.com";
		RegisterRequest adminRegisterRequest =
				new RegisterRequest(adminHandle, adminEmail, password, "Admin", "User", UserRole.ADMIN);
		ResponseEntity<AuthResponse> adminRegisterResponse =
				restTemplate.postForEntity(BASE_URL + "/auth/register", adminRegisterRequest, AuthResponse.class);
		assertThat(adminRegisterResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertThat(adminRegisterResponse.getBody()).isNotNull();
		assertThat(adminRegisterResponse.getBody().token()).isNotBlank();
		adminToken = adminRegisterResponse.getBody().token();
	}

	@Test
	@Order(2)
	void onlyAdminCanCreateProblem() {
		assertThat(userToken).as("User token should not be null").isNotNull();

		HttpHeaders userHeaders = new HttpHeaders();
		userHeaders.setBearerAuth(userToken);
		HttpEntity<CreateProblemRequest> userCreateProblemRequest =
				new HttpEntity<>(CREATE_PROBLEM_REQUEST, userHeaders);
		ResponseEntity<CreateProblemResponse> userCreateProblemResponse = restTemplate.postForEntity(
				BASE_URL + "/problems", userCreateProblemRequest, CreateProblemResponse.class);

		assertThat(userCreateProblemResponse.getStatusCode()).isEqualTo(HttpStatus.FORBIDDEN);

		assertThat(adminToken).as("Admin token should not be null").isNotNull();

		HttpHeaders adminHeaders = new HttpHeaders();
		adminHeaders.setBearerAuth(adminToken);
		HttpEntity<CreateProblemRequest> adminCreateProblemRequest =
				new HttpEntity<>(CREATE_PROBLEM_REQUEST, adminHeaders);
		ResponseEntity<CreateProblemResponse> adminCreateProblemResponse = restTemplate.postForEntity(
				BASE_URL + "/problems", adminCreateProblemRequest, CreateProblemResponse.class);

		assertThat(adminCreateProblemResponse.getStatusCode()).isEqualTo(HttpStatus.CREATED);
		ProblemDetailsUi problemDetails = adminCreateProblemResponse.getBody().problemDetails();
		problemId = problemDetails.id();
		assertProblemDetails(problemDetails);

		// Create additional problems to test the list and get APIs
		for (int i = 2; i <= PROBLEM_COUNT; i++) {
			adminCreateProblemResponse = restTemplate.postForEntity(
					BASE_URL + "/problems", adminCreateProblemRequest, CreateProblemResponse.class);
			assertThat(adminCreateProblemResponse.getStatusCode()).isEqualTo(HttpStatus.CREATED);
		}
	}

	@Test
	@Order(3)
	void retrieveProblems_withoutJwtHeaders() {
		ResponseEntity<ListProblemsResponse> listResponse =
				restTemplate.getForEntity(BASE_URL + "/problems/list", ListProblemsResponse.class);

		assertThat(listResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		List<ProblemSummaryUi> problems = listResponse.getBody().problems();
		// 20 is the page size
		assertThat(problems).isNotNull().hasSize(20);

		listResponse = restTemplate.getForEntity(BASE_URL + "/problems/list?page=2", ListProblemsResponse.class);

		assertThat(listResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		problems = listResponse.getBody().problems();
		assertThat(problems).isNotNull().hasSize(10);

		ResponseEntity<GetProblemByIdResponse> problemDetailsResponse =
				restTemplate.getForEntity(BASE_URL + "/problems/" + problemId, GetProblemByIdResponse.class);

		assertThat(problemDetailsResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertProblemDetails(problemDetailsResponse.getBody().problemDetails());
	}

	@Test
	@Order(4)
	void onlyLoggedInUsersCanMakeSubmissions() {
		SubmitCodeRequest submitCodeRequest = new SubmitCodeRequest(problemId, CODE, SUBMISSION_LANGUAGE);
		ResponseEntity<SubmitCodeResponse> submitCodeResponse =
				restTemplate.postForEntity(BASE_URL + "/submissions", submitCodeRequest, SubmitCodeResponse.class);

		assertThat(submitCodeResponse.getStatusCode()).isEqualTo(HttpStatus.FORBIDDEN);

		HttpHeaders userHeaders = new HttpHeaders();
		userHeaders.setBearerAuth(userToken);
		HttpEntity<SubmitCodeRequest> userSubmitCodeRequest = new HttpEntity<>(submitCodeRequest, userHeaders);
		ResponseEntity<SubmitCodeResponse> userSubmitCodeResponse =
				restTemplate.postForEntity(BASE_URL + "/submissions", userSubmitCodeRequest, SubmitCodeResponse.class);

		assertThat(userSubmitCodeResponse.getStatusCode()).isEqualTo(HttpStatus.CREATED);
		SubmissionDetailsUi submissionDetails = userSubmitCodeResponse.getBody().submissionDetails();
		assertSubmissionDetails(submissionDetails);
		submissionId = submissionDetails.id();

		// Create additional submissions to test the list and get APIs
		for (int i = 2; i <= SUBMISSION_COUNT; i++) {
			userSubmitCodeResponse = restTemplate.postForEntity(
					BASE_URL + "/submissions", userSubmitCodeRequest, SubmitCodeResponse.class);
			assertThat(userSubmitCodeResponse.getStatusCode()).isEqualTo(HttpStatus.CREATED);
		}
	}

	@Test
	@Order(5)
	void retrieveSubmissions_withoutJwtHeaders() {
		ResponseEntity<ListSubmissionsResponse> listResponse =
				restTemplate.getForEntity(BASE_URL + "/submissions/list", ListSubmissionsResponse.class);

		assertThat(listResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		List<SubmissionSummaryUi> submissions = listResponse.getBody().submissions();
		// 20 is the page size
		assertThat(submissions).isNotNull().hasSize(20);

		listResponse = restTemplate.getForEntity(BASE_URL + "/submissions/list?page=2", ListSubmissionsResponse.class);

		assertThat(listResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		submissions = listResponse.getBody().submissions();
		assertThat(submissions).isNotNull().hasSize(5);

		ResponseEntity<GetSubmissionByIdResponse> submissionDetailsResponse =
				restTemplate.getForEntity(BASE_URL + "/submissions/" + submissionId, GetSubmissionByIdResponse.class);

		assertThat(submissionDetailsResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertSubmissionDetails(submissionDetailsResponse.getBody().submissionDetails());
	}

	@Test
	@Order(6)
	void logoutUsers() {
		HttpHeaders userHeaders = new HttpHeaders();
		userHeaders.setBearerAuth(userToken);
		HttpEntity<Void> userLogoutRequest = new HttpEntity<>(userHeaders);
		ResponseEntity<Void> userLogoutResponse =
				restTemplate.postForEntity(BASE_URL + "/auth/logout", userLogoutRequest, Void.class);

		assertThat(userLogoutResponse.getStatusCode()).isEqualTo(HttpStatus.OK);

		HttpHeaders adminHeaders = new HttpHeaders();
		adminHeaders.setBearerAuth(adminToken);
		HttpEntity<Void> adminLogoutRequest = new HttpEntity<>(adminHeaders);
		ResponseEntity<Void> adminLogoutResponse =
				restTemplate.postForEntity(BASE_URL + "/auth/logout", adminLogoutRequest, Void.class);

		assertThat(adminLogoutResponse.getStatusCode()).isEqualTo(HttpStatus.OK);
	}

	private void assertProblemDetails(ProblemDetailsUi problemDetails) {
		assertThat(problemDetails).isNotNull();
		assertThat(problemDetails.title()).isEqualTo(PROBLEM_TITLE);
		assertThat(problemDetails.statement()).isEqualTo(PROBLEM_STATEMENT);
		assertThat(problemDetails.difficulty()).isEqualTo(PROBLEM_DIFFICULTY);
		assertThat(problemDetails.timeLimitInSecond()).isEqualTo(PROBLEM_TIME_LIMIT);
		assertThat(problemDetails.memoryLimitInMb()).isEqualTo(PROBLEM_MEMORY_LIMIT);
		assertThat(problemDetails.tags()).containsExactlyInAnyOrderElementsOf(PROBLEM_TAGS);
		assertThat(problemDetails.sampleTestCases()).hasSize(1);
	}

	private void assertProblemSummary(ProblemSummaryUi problemSummary) {
		assertThat(problemSummary).isNotNull();
		assertThat(problemSummary.title()).isEqualTo(PROBLEM_TITLE);
		assertThat(problemSummary.difficulty()).isEqualTo(PROBLEM_DIFFICULTY);
		assertThat(problemSummary.tags()).containsExactlyInAnyOrderElementsOf(PROBLEM_TAGS);
	}

	private void assertSubmissionDetails(SubmissionDetailsUi submissionDetails) {
		assertThat(submissionDetails).isNotNull();
		assertProblemSummary(submissionDetails.problemSummary());
		assertUserSummary(submissionDetails.userSummary());
		assertThat(submissionDetails.userSummary()).isNotNull();
		assertThat(submissionDetails.status()).isEqualTo(SubmissionStatus.WAITING_FOR_EXECUTION);
		assertThat(submissionDetails.language()).isEqualTo(SUBMISSION_LANGUAGE);
		assertThat(submissionDetails.code()).isEqualTo(CODE);
	}

	private void assertUserSummary(UserSummaryUi userSummary) {
		assertThat(userSummary).isNotNull();
		assertThat(userSummary.handle()).isEqualTo(userHandle);
	}
}
