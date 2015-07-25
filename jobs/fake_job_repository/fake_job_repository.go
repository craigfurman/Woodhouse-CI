// This file was generated by counterfeiter
package fake_job_repository

import (
	"sync"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type FakeRepository struct {
	SaveStub        func(job *jobs.Job) error
	saveMutex       sync.RWMutex
	saveArgsForCall []struct {
		job *jobs.Job
	}
	saveReturns struct {
		result1 error
	}
	FindByIdStub        func(id string) (jobs.Job, error)
	findByIdMutex       sync.RWMutex
	findByIdArgsForCall []struct {
		id string
	}
	findByIdReturns struct {
		result1 jobs.Job
		result2 error
	}
}

func (fake *FakeRepository) Save(job *jobs.Job) error {
	fake.saveMutex.Lock()
	fake.saveArgsForCall = append(fake.saveArgsForCall, struct {
		job *jobs.Job
	}{job})
	fake.saveMutex.Unlock()
	if fake.SaveStub != nil {
		return fake.SaveStub(job)
	} else {
		return fake.saveReturns.result1
	}
}

func (fake *FakeRepository) SaveCallCount() int {
	fake.saveMutex.RLock()
	defer fake.saveMutex.RUnlock()
	return len(fake.saveArgsForCall)
}

func (fake *FakeRepository) SaveArgsForCall(i int) *jobs.Job {
	fake.saveMutex.RLock()
	defer fake.saveMutex.RUnlock()
	return fake.saveArgsForCall[i].job
}

func (fake *FakeRepository) SaveReturns(result1 error) {
	fake.SaveStub = nil
	fake.saveReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) FindById(id string) (jobs.Job, error) {
	fake.findByIdMutex.Lock()
	fake.findByIdArgsForCall = append(fake.findByIdArgsForCall, struct {
		id string
	}{id})
	fake.findByIdMutex.Unlock()
	if fake.FindByIdStub != nil {
		return fake.FindByIdStub(id)
	} else {
		return fake.findByIdReturns.result1, fake.findByIdReturns.result2
	}
}

func (fake *FakeRepository) FindByIdCallCount() int {
	fake.findByIdMutex.RLock()
	defer fake.findByIdMutex.RUnlock()
	return len(fake.findByIdArgsForCall)
}

func (fake *FakeRepository) FindByIdArgsForCall(i int) string {
	fake.findByIdMutex.RLock()
	defer fake.findByIdMutex.RUnlock()
	return fake.findByIdArgsForCall[i].id
}

func (fake *FakeRepository) FindByIdReturns(result1 jobs.Job, result2 error) {
	fake.FindByIdStub = nil
	fake.findByIdReturns = struct {
		result1 jobs.Job
		result2 error
	}{result1, result2}
}

var _ jobs.Repository = new(FakeRepository)