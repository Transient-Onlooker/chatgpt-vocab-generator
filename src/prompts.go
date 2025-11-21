package main

import (
	"fmt"
	"strings"
)

func buildPrompts(parsed []VocabPair, questionType string, numSentences int) (systemPrompt, userPrompt string) {
	distributionRule := "2. CRITICAL: The position of the correct answer MUST be truly and unpredictably randomized to ensure a balanced distribution. For the entire set of questions, each choice position (①, ②, ③, ④, ⑤) should be the correct answer approximately 20% of the time. DO NOT use any discernible pattern (e.g., 1, 2, 3, 4, 5 or 5, 4, 3, 2, 1). The sequence of correct answers must appear random and chaotic."
	selfCorrectionRule := "### Final Review\nBefore concluding your response, you MUST review the entire generated text one last time to ensure every single rule has been followed. Pay special attention that every question has exactly 5 numbered choices (① to ⑤). If you find any mistake, you must correct it before finishing."

	var systemPromptLines []string
	switch questionType {
	case "빈칸 추론":
		systemPromptLines = []string{
			"You are an expert English vocabulary test maker for Korean students.",
			"Your task is to create multiple-choice questions that test understanding of words in context.",
			"Strictly follow all rules below.",
			"",
			"### Main Rule",
			"For each WORD and for each of its SENSEs, you must generate a complete question block.",
			"",
			"### Word Selection & Question Style Rule",
			"1. PRIORITY: Focus on polysemous words—those with multiple, distinct meanings (e.g., different parts of speech like 'conduct' as a noun vs. verb, or different senses like 'bank' of a river vs. a financial institution).",
			"2. GOAL: The questions should be intentionally challenging, designed to confuse the test-taker and test their ability to discern the correct meaning from context.",
			"",
			"### Answer Generation Rules",
			"1. CRITICAL: DO NOT mark the correct answer in the choices. Instead, create a separate `[정답]` section at the very end of the entire output, listing each question number and its correct choice number.",
			distributionRule,
			"",
			"### Output Structure (per question)",
			"1. Start with the question number (e.g., '1.').",
			"2. Add the title: '다음 빈칸에 공통으로 들어갈 말로 가장 적절한 것은?'",
			fmt.Sprintf("3. Provide exactly %d distinct English sentences as context. Each sentence must have the word blanked out as '_______'.", numSentences),
			"4. Provide exactly 5 answer choices (①, ②, ③, ④, ⑤).",
			"5. The choices must include one correct answer (the original WORD) and four plausible but incorrect distractors.",
			"6. Separate each full question block with a '---' line.",
			"",
			selfCorrectionRule,
		}
	case "영영풀이":
		systemPromptLines = []string{
			"You are an expert English vocabulary test maker for Korean students.",
			"Your task is to create multiple-choice questions based on English definitions.",
			"Strictly follow all rules below.",
			"",
			"### Main Rule",
			"For each WORD, you must generate one complete multiple-choice question.",
			"",
			"### Word Selection & Question Style Rule",
			"1. PRIORITY: Focus on polysemous words—those with multiple, distinct meanings (e.g., different parts of speech like 'conduct' as a noun vs. verb, or different senses like 'bank' of a river vs. a financial institution).",
			"2. GOAL: The questions should be intentionally challenging, designed to confuse the test-taker and test their ability to discern the correct meaning from context.",
			"",
			"### Answer Generation Rules",
			"1. CRITICAL: DO NOT mark the correct answer in the choices. Instead, create a separate `[정답]` section at the very end of the entire output, listing each question number and its correct choice number.",
			distributionRule,
			"",
			"### Output Structure (per question)",
			"1. Start with the question number (e.g., '1.').",
			"2. Add the title: '다음 영어 설명에 해당하는 단어는?'",
			"3. Provide the English definition of the WORD as the question body.",
			"4. Provide exactly 5 answer choices (①, ②, ③, ④, ⑤): one correct answer (the original WORD) and four plausible distractors (e.g., synonyms, related words).",
			"5. Separate each full question block with a '---' line.",
			"",
			selfCorrectionRule,
		}
	case "뜻풀이 판단":
		systemPromptLines = []string{
			"You are an expert English vocabulary test maker for Korean students.",
			"Your task is to create multiple-choice questions that test the precise definition of a word.",
			"Strictly follow all rules below.",
			"",
			"### Main Rule",
			"For each WORD, you must generate one complete multiple-choice question asking for its correct definition.",
			"",
			"### Word Selection & Question Style Rule",
			"1. PRIORITY: Focus on polysemous words—those with multiple, distinct meanings (e.g., different parts of speech like 'conduct' as a noun vs. verb, or different senses like 'bank' of a river vs. a financial institution).",
			"2. GOAL: The questions should be intentionally challenging, designed to confuse the test-taker and test their ability to discern the correct meaning from context.",
			"",
			"### Answer Generation Rules",
			"1. CRITICAL: DO NOT mark the correct answer in the choices. Instead, create a separate `[정답]` section at the very end of the entire output, listing each question number and its correct choice number.",
			distributionRule,
			"",
			"### Output Structure (per question)",
			"1. Start with the question number (e.g., '1.').",
			"2. Add the title: '다음 단어 <WORD>의 영영풀이로 가장 적절한 것은?' (replace <WORD> with the actual word).",
			"3. Provide exactly 5 definition choices (①, ②, ③, ④, ⑤): one perfectly correct definition and four subtly incorrect but plausible definitions.",
			"4. Separate each full question block with a '---' line.",
			"",
			selfCorrectionRule,
		}
	}
	systemPrompt = strings.Join(systemPromptLines, "\n")

	var parsedForModel []string
	for _, p := range parsed {
		parsedForModel = append(parsedForModel, fmt.Sprintf("%s = %s", p.Word, strings.Join(p.Meanings, ", ")))
	}
	parsedForModelText := strings.Join(parsedForModel, "\n")

	userPromptLines := []string{
		"Here is the list of vocabulary. Create test questions based on these words, strictly following all rules defined in the system instructions.",
		"",
		"[Vocabulary List]",
		parsedForModelText,
	}
	userPrompt = strings.Join(userPromptLines, "\n")

	return 
}
