// App.vue
<template>
    <div id="app">
        <header>
            <h1>Transaction Reconciliation System</h1>
        </header>

        <div class="container">
            <div class="upload-section">
                <h2>Upload Transaction Files</h2>
                <form @submit.prevent="submitFiles" class="upload-form">
                    <div class="file-inputs">
                        <div class="file-input">
                            <label for="fileA">Source System A:</label>
                            <input type="file" id="fileA" ref="fileA" accept=".csv" @change="onFileASelected"
                                required />
                            <span v-if="fileAName" class="file-name">{{ fileAName }}</span>
                        </div>

                        <div class="file-input">
                            <label for="fileB">Source System B:</label>
                            <input type="file" id="fileB" ref="fileB" accept=".csv" @change="onFileBSelected"
                                required />
                            <span v-if="fileBName" class="file-name">{{ fileBName }}</span>
                        </div>
                    </div>

                    <button type="submit" class="upload-button"
                        :disabled="isLoading || !fileASelected || !fileBSelected">
                        <span v-if="isLoading">Processing...</span>
                        <span v-else>Reconcile Files</span>
                    </button>
                </form>
            </div>

            <div v-if="isLoading" class="loading">
                <div class="spinner"></div>
                <p>Processing files. This may take a while for large files...</p>
            </div>

            <div v-if="errorMessage" class="error-message">
                <p>{{ errorMessage }}</p>
            </div>

            <div v-if="reconciliationResults" class="results-section">
                <h2>Reconciliation Results</h2>

                <div class="summary-card">
                    <h3>Summary</h3>
                    <div class="summary-stats">
                        <div class="stat-item">
                            <span class="stat-label">Files Processed:</span>
                            <span class="stat-value">System A ({{ reconciliationResults.summary.total_transactions_a }}
                                transactions) and System B ({{ reconciliationResults.summary.total_transactions_b }}
                                transactions)</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Total Discrepancies:</span>
                            <span class="stat-value">{{ reconciliationResults.summary.total_discrepancies }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Missing in System A:</span>
                            <span class="stat-value">{{ reconciliationResults.summary.missing_in_a_count }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Missing in System B:</span>
                            <span class="stat-value">{{ reconciliationResults.summary.missing_in_b_count }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Amount Mismatches:</span>
                            <span class="stat-value">{{ reconciliationResults.summary.amount_mismatches_count }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Status Mismatches:</span>
                            <span class="stat-value">{{ reconciliationResults.summary.status_mismatches_count }}</span>
                        </div>
                    </div>
                </div>

                <div class="tabs">
                    <button @click="activeTab = 'missingA'" :class="{ active: activeTab === 'missingA' }">
                        Missing in A ({{ reconciliationResults.missing_in_a.length }})
                    </button>
                    <button @click="activeTab = 'missingB'" :class="{ active: activeTab === 'missingB' }">
                        Missing in B ({{ reconciliationResults.missing_in_b.length }})
                    </button>
                    <button @click="activeTab = 'amountMismatches'"
                        :class="{ active: activeTab === 'amountMismatches' }">
                        Amount Mismatches ({{ reconciliationResults.amount_mismatches.length }})
                    </button>
                    <button @click="activeTab = 'statusMismatches'"
                        :class="{ active: activeTab === 'statusMismatches' }">
                        Status Mismatches ({{ reconciliationResults.status_mismatches.length }})
                    </button>
                </div>

                <div class="tab-content">
                    <!-- Missing in A Tab -->
                    <div v-if="activeTab === 'missingA'">
                        <h3>Transactions in System B but Missing in System A</h3>
                        <div v-if="reconciliationResults.missing_in_a.length === 0" class="no-results">
                            No transactions missing in System A
                        </div>
                        <div v-else class="results-table-container">
                            <table class="results-table">
                                <thead>
                                    <tr>
                                        <th>Transaction ID</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(id, index) in paginatedMissingA" :key="`missing-a-${index}`">
                                        <td>{{ id }}</td>
                                    </tr>
                                </tbody>
                            </table>
                            <div class="pagination" v-if="reconciliationResults.missing_in_a.length > pageSize">
                                <button @click="currentPage.missingA--" :disabled="currentPage.missingA === 1">
                                    Previous
                                </button>
                                <span>Page {{ currentPage.missingA }} of {{ totalPages.missingA }}</span>
                                <button @click="currentPage.missingA++"
                                    :disabled="currentPage.missingA >= totalPages.missingA">
                                    Next
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Missing in B Tab -->
                    <div v-if="activeTab === 'missingB'">
                        <h3>Transactions in System A but Missing in System B</h3>
                        <div v-if="reconciliationResults.missing_in_b.length === 0" class="no-results">
                            No transactions missing in System B
                        </div>
                        <div v-else class="results-table-container">
                            <table class="results-table">
                                <thead>
                                    <tr>
                                        <th>Transaction ID</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(id, index) in paginatedMissingB" :key="`missing-b-${index}`">
                                        <td>{{ id }}</td>
                                    </tr>
                                </tbody>
                            </table>
                            <div class="pagination" v-if="reconciliationResults.missing_in_b.length > pageSize">
                                <button @click="currentPage.missingB--" :disabled="currentPage.missingB === 1">
                                    Previous
                                </button>
                                <span>Page {{ currentPage.missingB }} of {{ totalPages.missingB }}</span>
                                <button @click="currentPage.missingB++"
                                    :disabled="currentPage.missingB >= totalPages.missingB">
                                    Next
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Amount Mismatches Tab -->
                    <div v-if="activeTab === 'amountMismatches'">
                        <h3>Amount Mismatches Between Systems</h3>
                        <div v-if="reconciliationResults.amount_mismatches.length === 0" class="no-results">
                            No amount mismatches found
                        </div>
                        <div v-else class="results-table-container">
                            <table class="results-table">
                                <thead>
                                    <tr>
                                        <th>Transaction ID</th>
                                        <th>Amount in A</th>
                                        <th>Amount in B</th>
                                        <th>Currency</th>
                                        <th>Difference</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(mismatch, index) in paginatedAmountMismatches" :key="`amount-${index}`">
                                        <td>{{ mismatch.transaction_id }}</td>
                                        <td>{{ mismatch.amount_a.toFixed(2) }}</td>
                                        <td>{{ mismatch.amount_b.toFixed(2) }}</td>
                                        <td>{{ mismatch.currency }}</td>
                                        <td>{{ (mismatch.amount_a - mismatch.amount_b).toFixed(2) }}</td>
                                    </tr>
                                </tbody>
                            </table>
                            <div class="pagination" v-if="reconciliationResults.amount_mismatches.length > pageSize">
                                <button @click="currentPage.amountMismatches--"
                                    :disabled="currentPage.amountMismatches === 1">
                                    Previous
                                </button>
                                <span>Page {{ currentPage.amountMismatches }} of {{ totalPages.amountMismatches
                                    }}</span>
                                <button @click="currentPage.amountMismatches++"
                                    :disabled="currentPage.amountMismatches >= totalPages.amountMismatches">
                                    Next
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Status Mismatches Tab -->
                    <div v-if="activeTab === 'statusMismatches'">
                        <h3>Status Mismatches Between Systems</h3>
                        <div v-if="reconciliationResults.status_mismatches.length === 0" class="no-results">
                            No status mismatches found
                        </div>
                        <div v-else class="results-table-container">
                            <table class="results-table">
                                <thead>
                                    <tr>
                                        <th>Transaction ID</th>
                                        <th>Status in A</th>
                                        <th>Status in B</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(mismatch, index) in paginatedStatusMismatches" :key="`status-${index}`">
                                        <td>{{ mismatch.transaction_id }}</td>
                                        <td>{{ mismatch.status_a }}</td>
                                        <td>{{ mismatch.status_b }}</td>
                                    </tr>
                                </tbody>
                            </table>
                            <div class="pagination" v-if="reconciliationResults.status_mismatches.length > pageSize">
                                <button @click="currentPage.statusMismatches--"
                                    :disabled="currentPage.statusMismatches === 1">
                                    Previous
                                </button>
                                <span>Page {{ currentPage.statusMismatches }} of {{ totalPages.statusMismatches
                                    }}</span>
                                <button @click="currentPage.statusMismatches++"
                                    :disabled="currentPage.statusMismatches >= totalPages.statusMismatches">
                                    Next
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: 'App',
    data() {
        return {
            fileASelected: false,
            fileBSelected: false,
            fileAName: '',
            fileBName: '',
            isLoading: false,
            errorMessage: '',
            reconciliationResults: null,
            activeTab: 'missingA',
            pageSize: 25,
            currentPage: {
                missingA: 1,
                missingB: 1,
                amountMismatches: 1,
                statusMismatches: 1
            }
        };
    },
    computed: {
        totalPages() {
            if (!this.reconciliationResults) return { missingA: 0, missingB: 0, amountMismatches: 0, statusMismatches: 0 };

            return {
                missingA: Math.ceil(this.reconciliationResults.missing_in_a.length / this.pageSize),
                missingB: Math.ceil(this.reconciliationResults.missing_in_b.length / this.pageSize),
                amountMismatches: Math.ceil(this.reconciliationResults.amount_mismatches.length / this.pageSize),
                statusMismatches: Math.ceil(this.reconciliationResults.status_mismatches.length / this.pageSize)
            };
        },
        paginatedMissingA() {
            if (!this.reconciliationResults) return [];

            const start = (this.currentPage.missingA - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.reconciliationResults.missing_in_a.slice(start, end);
        },
        paginatedMissingB() {
            if (!this.reconciliationResults) return [];

            const start = (this.currentPage.missingB - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.reconciliationResults.missing_in_b.slice(start, end);
        },
        paginatedAmountMismatches() {
            if (!this.reconciliationResults) return [];

            const start = (this.currentPage.amountMismatches - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.reconciliationResults.amount_mismatches.slice(start, end);
        },
        paginatedStatusMismatches() {
            if (!this.reconciliationResults) return [];

            const start = (this.currentPage.statusMismatches - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.reconciliationResults.status_mismatches.slice(start, end);
        }
    },
    methods: {
        onFileASelected(event) {
            const file = event.target.files[0];
            if (file) {
                this.fileASelected = true;
                this.fileAName = file.name;
            } else {
                this.fileASelected = false;
                this.fileAName = '';
            }
        },
        onFileBSelected(event) {
            const file = event.target.files[0];
            if (file) {
                this.fileBSelected = true;
                this.fileBName = file.name;
            } else {
                this.fileBSelected = false;
                this.fileBName = '';
            }
        },
        async submitFiles() {
            if (!this.fileASelected || !this.fileBSelected) {
                this.errorMessage = 'Please select both files before submitting.';
                return;
            }

            this.isLoading = true;
            this.errorMessage = '';
            this.reconciliationResults = null;

            try {
                const fileA = this.$refs.fileA.files[0];
                const fileB = this.$refs.fileB.files[0];

                // Create FormData object
                const formData = new FormData();
                formData.append('fileA', fileA);
                formData.append('fileB', fileB);

                // Submit to backend API
                const response = await fetch('http://localhost:7080/api/reconcile', {
                    method: 'POST',
                    body: formData
                });

                if (!response.ok) {
                    throw new Error(`Server responded with status: ${response.status}`);
                }

                const data = await response.json();
                this.reconciliationResults = data;

                // Reset pagination to first page
                for (const key in this.currentPage) {
                    this.currentPage[key] = 1;
                }

                // Set active tab to the one with most issues
                const counts = [
                    { tab: 'missingA', count: data.missing_in_a.length },
                    { tab: 'missingB', count: data.missing_in_b.length },
                    { tab: 'amountMismatches', count: data.amount_mismatches.length },
                    { tab: 'statusMismatches', count: data.status_mismatches.length }
                ];

                const maxTab = counts.reduce((prev, current) =>
                    (prev.count > current.count) ? prev : current
                );

                this.activeTab = maxTab.tab;
            } catch (error) {
                console.error('Error processing files:', error);
                this.errorMessage = `Error processing files: ${error.message}`;
            } finally {
                this.isLoading = false;
            }
        }
    }
};
</script>

<style>
:root {
    --primary-color: #2c3e50;
    --secondary-color: #42b983;
    --error-color: #e74c3c;
    --light-gray: #f5f5f5;
    --medium-gray: #e0e0e0;
    --dark-gray: #999;
    --border-radius: 4px;
    --shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: 'Arial', sans-serif;
    line-height: 1.6;
    color: var(--primary-color);
    background-color: #f9f9f9;
}

#app {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    text-align: center;
    margin-bottom: 30px;
    padding: 20px 0;
    border-bottom: 1px solid var(--medium-gray);
}

h1 {
    color: var(--primary-color);
}

.container {
    display: flex;
    flex-direction: column;
    gap: 30px;
}

.upload-section {
    background-color: white;
    padding: 25px;
    border-radius: var(--border-radius);
    box-shadow: var(--shadow);
}

.upload-form {
    margin-top: 20px;
}

.file-inputs {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin-bottom: 20px;
}

.file-input {
    display: flex;
    flex-direction: column;
    gap: 5px;
}

.file-input label {
    font-weight: bold;
}

.file-input input[type="file"] {
    padding: 10px;
    border: 1px solid var(--medium-gray);
    border-radius: var(--border-radius);
    background-color: var(--light-gray);
    width: 100%;
}

.file-name {
    font-size: 14px;
    color: var(--dark-gray);
}

.upload-button {
    padding: 12px 24px;
    background-color: var(--secondary-color);
    color: white;
    border: none;
    border-radius: var(--border-radius);
    cursor: pointer;
    font-weight: bold;
    transition: background-color 0.3s;
    width: 100%;
}

.upload-button:hover {
    background-color: #3aa876;
}

.upload-button:disabled {
    background-color: var(--dark-gray);
    cursor: not-allowed;
}

.loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    background-color: white;
    border-radius: var(--border-radius);
    box-shadow: var(--shadow);
}

.spinner {
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top: 4px solid var(--secondary-color);
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin-bottom: 20px;
}

@keyframes spin {
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
}

.error-message {
    padding: 15px;
    background-color: #fde2e2;
    border-left: 4px solid var(--error-color);
    color: #a83232;
    border-radius: var(--border-radius);
}

.results-section {
    background-color: white;
    padding: 25px;
    border-radius: var(--border-radius);
    box-shadow: var(--shadow);
}

.summary-card {
    background-color: var(--light-gray);
    padding: 20px;
    border-radius: var(--border-radius);
    margin-top: 20px;
    margin-bottom: 30px;
}

.summary-stats {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 15px;
    margin-top: 15px;
}

.stat-item {
    display: flex;
    flex-direction: column;
}

.stat-label {
    font-weight: bold;
    font-size: 14px;
    color: var(--dark-gray);
}

.stat-value {
    font-size: 18px;
    color: var(--primary-color);
}

.tabs {
    display: flex;
    margin-bottom: 20px;
    overflow-x: auto;
    border-bottom: 1px solid var(--medium-gray);
}

.tabs button {
    padding: 12px 20px;
    background-color: transparent;
    border: none;
    cursor: pointer;
    font-weight: bold;
    color: var(--dark-gray);
    position: relative;
    white-space: nowrap;
}

.tabs button.active {
    color: var(--secondary-color);
}

.tabs button.active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    width: 100%;
    height: 3px;
    background-color: var(--secondary-color);
}

.tab-content {
    margin-top: 20px;
}

.no-results {
    padding: 30px;
    text-align: center;
    font-style: italic;
    color: var(--dark-gray);
    background-color: var(--light-gray);
    border-radius: var(--border-radius);
}

.results-table-container {
    overflow-x: auto;
}

.results-table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 20px;
}

.results-table th,
.results-table td {
    padding: 12px 15px;
    text-align: left;
    border-bottom: 1px solid var(--medium-gray);
}

.results-table th {
    background-color: var(--light-gray);
    font-weight: bold;
}

.results-table tr:hover {
    background-color: rgba(66, 185, 131, 0.05);
}

.pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 20px;
    margin-top: 20px;
}

.pagination button {
    padding: 8px 16px;
    background-color: var(--light-gray);
    border: 1px solid var(--medium-gray);
    border-radius: var(--border-radius);
    cursor: pointer;
    transition: background-color 0.3s;
}

.pagination button:hover:not(:disabled) {
    background-color: var(--medium-gray);
}

.pagination button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

@media (max-width: 768px) {
    .summary-stats {
        grid-template-columns: 1fr;
    }

    .tabs {
        flex-direction: column;
        border-bottom: none;
    }

    .tabs button {
        border-bottom: 1px solid var(--medium-gray);
    }

    .tabs button.active::after {
        display: none;
    }

    .tabs button.active {
        background-color: var(--light-gray);
    }
}
</style>