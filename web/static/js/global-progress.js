// Work Queue System - Sequential Job Processing
// Processes actions in order (FIFO) with visual progress feedback
// Persists queue state across page navigation

(function() {
    'use strict';

    const STORAGE_KEY = 'mediacheky_work_queue';
    const PROGRESS_KEY = 'mediacheky_current_progress';
    const CHECK_INTERVAL = 100; // Check every 100ms for responsiveness

    // Work Queue Manager - Sequential job processing
    window.WorkQueue = {
        queue: [],          // Pending jobs
        current: null,      // Currently executing job
        processing: false,  // Lock to prevent concurrent execution

        // Initialize the queue system
        init() {
            this.loadFromStorage();
            this.startProcessing();
            
            // Listen for storage changes from other tabs
            window.addEventListener('storage', (e) => {
                if (e.key === STORAGE_KEY || e.key === PROGRESS_KEY) {
                    this.loadFromStorage();
                }
            });

            console.log('[WorkQueue] Initialized');
        },

        // Load state from localStorage
        loadFromStorage() {
            try {
                const storedQueue = localStorage.getItem(STORAGE_KEY);
                const storedProgress = localStorage.getItem(PROGRESS_KEY);
                
                if (storedQueue) {
                    this.queue = JSON.parse(storedQueue);
                }
                if (storedProgress) {
                    this.current = JSON.parse(storedProgress);
                }
            } catch (e) {
                console.error('[WorkQueue] Failed to load from storage:', e);
            }
        },

        // Save state to localStorage
        saveToStorage() {
            try {
                localStorage.setItem(STORAGE_KEY, JSON.stringify(this.queue));
                if (this.current) {
                    localStorage.setItem(PROGRESS_KEY, JSON.stringify(this.current));
                } else {
                    localStorage.removeItem(PROGRESS_KEY);
                }
            } catch (e) {
                console.error('[WorkQueue] Failed to save to storage:', e);
            }
        },

        // Add a job to the queue
        enqueue(job) {
            const jobWithId = {
                ...job,
                id: job.id || `job-${Date.now()}-${Math.random()}`,
                addedAt: Date.now(),
                status: 'pending'
            };
            
            this.queue.push(jobWithId);
            this.saveToStorage();
            this.notifyProgressListeners(); // Notify immediately on enqueue
            console.log(`[WorkQueue] Enqueued job: ${jobWithId.id}`, jobWithId);
            
            // Start processing if not already running
            if (!this.processing) {
                this.processNext();
            }
            
            return jobWithId.id;
        },

        // Process next job in queue
        async processNext() {
            if (this.processing || this.queue.length === 0) {
                return;
            }

            this.processing = true;
            const job = this.queue.shift();
            this.current = {
                ...job,
                status: 'running',
                startTime: Date.now(),
                currentStep: 0
            };
            
            this.saveToStorage();
            this.notifyProgressListeners();
            
            console.log(`[WorkQueue] Processing job: ${job.id}`, job);

            try {
                // Execute the job function
                if (typeof job.execute === 'function') {
                    await job.execute(this);
                } else {
                    console.error('[WorkQueue] Job has no execute function', job);
                }
                
                console.log(`[WorkQueue] Completed job: ${job.id}`);
            } catch (error) {
                console.error(`[WorkQueue] Job failed: ${job.id}`, error);
                // Could add retry logic here
            } finally {
                this.current = null;
                this.processing = false;
                this.saveToStorage();
                this.notifyProgressListeners();
                
                // Process next job if any
                if (this.queue.length > 0) {
                    setTimeout(() => this.processNext(), 100);
                }
            }
        },

        // Start processing loop
        startProcessing() {
            setInterval(() => {
                // Auto-cleanup stale jobs (> 10 minutes old)
                const now = Date.now();
                this.queue = this.queue.filter(job => {
                    const age = now - job.addedAt;
                    if (age > 10 * 60 * 1000) {
                        console.warn('[WorkQueue] Removing stale job:', job.id);
                        return false;
                    }
                    return true;
                });

                // Cleanup stale current job
                if (this.current && (now - this.current.startTime) > 10 * 60 * 1000) {
                    console.warn('[WorkQueue] Cleaning up stale current job');
                    this.current = null;
                    this.processing = false;
                    this.saveToStorage();
                    this.notifyProgressListeners();
                }

                // Try to process next if idle
                if (!this.processing && this.queue.length > 0) {
                    this.processNext();
                }
            }, CHECK_INTERVAL);
        },

        // Update current job step
        updateStep(stepIndex = null) {
            if (!this.current) return;
            
            if (stepIndex !== null) {
                this.current.currentStep = stepIndex;
            } else {
                this.current.currentStep++;
            }
            
            if (this.current.currentStep >= this.current.steps.length) {
                this.current.currentStep = this.current.steps.length - 1;
            }
            
            this.saveToStorage();
            this.notifyProgressListeners();
        },

        // Get current progress for display
        getCurrentProgress() {
            return this.current;
        },

        // Get queue status
        getStatus() {
            return {
                queueLength: this.queue.length,
                current: this.current,
                processing: this.processing
            };
        },

        // Get all jobs for display (current + pending)
        getAllJobs() {
            const jobs = [];
            
            // Add current job if exists
            if (this.current) {
                jobs.push({
                    ...this.current,
                    status: 'running'
                });
            }
            
            // Add pending jobs
            jobs.push(...this.queue.map(job => ({
                ...job,
                status: 'pending'
            })));
            
            return jobs;
        },

        // Progress listeners
        progressListeners: [],

        addProgressListener(callback) {
            this.progressListeners.push(callback);
            // Call immediately with current state
            callback(this.current);
        },

        removeProgressListener(callback) {
            this.progressListeners = this.progressListeners.filter(l => l !== callback);
        },

        notifyProgressListeners() {
            this.progressListeners.forEach(callback => {
                try {
                    callback(this.current);
                } catch (e) {
                    console.error('[WorkQueue] Progress listener error:', e);
                }
            });
        }
    };

    // Auto-initialize
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.WorkQueue.init();
        });
    } else {
        window.WorkQueue.init();
    }

    // Register Alpine.js component BEFORE Alpine initializes
    document.addEventListener('alpine:init', () => {
        console.log('[WorkQueue] Registering Alpine.js component');
        
        Alpine.data('globalProgress', () => ({
            showProgress: false,
            allJobs: [],

            init() {
                console.log('[globalProgress] Component initialized');
                
                const updateProgress = () => {
                    const jobs = window.WorkQueue.getAllJobs();
                    this.allJobs = jobs;
                    this.showProgress = jobs.length > 0;
                    
                    console.log('[globalProgress] Updated jobs:', jobs.length);
                };

                // Initial state
                updateProgress();

                // Listen for updates (called on enqueue, step changes, completion)
                window.WorkQueue.addProgressListener(updateProgress);
                
                // Poll for changes every 500ms as backup
                setInterval(updateProgress, 500);
            },

            getJobIcon(status) {
                if (status === 'completed') return '✅';
                if (status === 'running') return '⏳';
                return '⏸️';
            },

            getStepIcon(job, stepIndex) {
                if (job.status === 'pending') return '⏸️';
                if (job.status === 'completed') return '✅';
                // Running job
                if (stepIndex < job.currentStep) return '✅';
                if (stepIndex === job.currentStep) return '⏳';
                return '⏸️';
            },

            getStepStatus(job, stepIndex) {
                if (job.status === 'pending') return 'PENDING';
                if (job.status === 'completed') return 'OK';
                // Running job
                if (stepIndex < job.currentStep) return 'OK';
                if (stepIndex === job.currentStep) return 'RUNNING';
                return 'PENDING';
            }
        }));
    });
})();
