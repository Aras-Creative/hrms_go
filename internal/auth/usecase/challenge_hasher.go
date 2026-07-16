package usecase

type ChallengeHasher interface {
	HashChallenge(challenge string) (string, error)
	VerifyChallenge(hash, challenge string) error
}
