# Bài tập: Data Preprocessing

## 📝 Đề bài

Bạn có dataset `student_performance.csv` với thông tin học sinh:

```csv
StudentID,Name,Gender,Age,StudyHours,PreviousScore,AttendanceRate,Passed
1,Alice,Female,20,5.5,85,90,1
2,Bob,Male,19,,75,80,1
3,Charlie,Male,21,3.2,NaN,70,0
4,Diana,Female,20,6.8,90,95,1
5,Eve,Female,22,2.1,65,NaN,0
```

**Nhiệm vụ**: Preprocessing data để train model dự đoán `Passed` (0/1)

**Yêu cầu**:
1. Load data và phân tích
2. Xử lý missing values
3. Encode categorical features (Gender)
4. Feature scaling
5. Split train/test (80/20)
6. In ra shape của X_train, X_test

---

## 💡 Approach/Solution

### Phân tích vấn đề
- **Features**: Gender (categorical), Age, StudyHours, PreviousScore, AttendanceRate (numerical)
- **Target**: Passed (binary)
- **Missing values**:
  - StudyHours: Bob (row 2)
  - PreviousScore: Charlie (row 3)
  - AttendanceRate: Eve (row 5)

### Strategy
1. **Drop columns** không cần: StudentID, Name
2. **Impute missing values** bằng mean
3. **Label Encoding** cho Gender (Male=0, Female=1)
4. **Feature Scaling**: StandardScaler
5. **Split**: train_test_split với test_size=0.2

---

## 🔧 Hướng dẫn implement chi tiết

### Bước 1: Setup và load data

```python
# 1. Import libraries
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.impute import SimpleImputer

# 2. Create sample data (hoặc load từ CSV)
data = {
    'StudentID': [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    'Name': ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack'],
    'Gender': ['Female', 'Male', 'Male', 'Female', 'Female', 'Male', 'Female', 'Male', 'Female', 'Male'],
    'Age': [20, 19, 21, 20, 22, 19, 21, 20, 19, 22],
    'StudyHours': [5.5, np.nan, 3.2, 6.8, 2.1, 4.5, 7.2, 3.8, 5.0, np.nan],
    'PreviousScore': [85, 75, np.nan, 90, 65, 80, 95, 70, 88, 78],
    'AttendanceRate': [90, 80, 70, 95, np.nan, 85, 92, 75, np.nan, 88],
    'Passed': [1, 1, 0, 1, 0, 1, 1, 0, 1, 1]
}
dataset = pd.DataFrame(data)

# Hoặc load từ file
# dataset = pd.read_csv('student_performance.csv')

print("Dataset shape:", dataset.shape)
print("\nFirst 5 rows:")
print(dataset.head())
print("\nData info:")
print(dataset.info())
print("\nMissing values:")
print(dataset.isnull().sum())
```

### Bước 2: Drop unnecessary columns

```python
# Drop StudentID và Name (không phải features)
dataset = dataset.drop(['StudentID', 'Name'], axis=1)

# Separate features và target
X = dataset.drop('Passed', axis=1).values  # Tất cả columns trừ Passed
y = dataset['Passed'].values               # Cột Passed

print("\nX shape before preprocessing:", X.shape)  # (10, 5)
print("y shape:", y.shape)  # (10,)
```

**Giải thích**:
- `axis=1`: drop theo columns
- `.values`: convert DataFrame → NumPy array

### Bước 3: Handle missing values

```python
# Identify numerical columns (tất cả trừ Gender - column 0)
# Column 0: Gender (categorical)
# Columns 1-4: Age, StudyHours, PreviousScore, AttendanceRate (numerical)

# Impute missing values bằng mean cho numerical columns
imputer = SimpleImputer(missing_values=np.nan, strategy='mean')
X[:, 1:5] = imputer.fit_transform(X[:, 1:5])  # Columns 1,2,3,4

print("\nAfter imputation:")
print(X)
```

**Giải thích**:
- `X[:, 1:5]`: lấy all rows, columns từ 1 đến 4
- `strategy='mean'`: thay NaN bằng mean
- **Alternative strategies**: `'median'`, `'most_frequent'`

### Bước 4: Encode categorical feature

```python
# Gender column (index 0) - Label Encoding
le = LabelEncoder()
X[:, 0] = le.fit_transform(X[:, 0])
# Female → 0 (hoặc 1 tùy alphabetical order)
# Male → 1 (hoặc 0)

print("\nAfter encoding:")
print(X[:5])  # Print first 5 rows
print("\nGender mapping:", dict(zip(le.classes_, le.transform(le.classes_))))
```

**Giải thích**:
- LabelEncoder cho binary categorical (2 values)
- Output: Female=0, Male=1 (hoặc ngược lại)

### Bước 5: Split train/test

```python
# Split 80% train, 20% test
X_train, X_test, y_train, y_test = train_test_split(
    X, y,
    test_size=0.2,    # 20% test
    random_state=42   # Reproducible
)

print("\nTrain/Test split:")
print(f"X_train shape: {X_train.shape}")  # (8, 5)
print(f"X_test shape: {X_test.shape}")    # (2, 5)
print(f"y_train shape: {y_train.shape}")  # (8,)
print(f"y_test shape: {y_test.shape}")    # (2,)
```

### Bước 6: Feature Scaling

```python
# StandardScaler: (x - mean) / std
sc = StandardScaler()
X_train = sc.fit_transform(X_train)  # Fit trên train, transform train
X_test = sc.transform(X_test)        # Transform test với mean/std từ train

print("\nX_train (scaled) - first 3 rows:")
print(X_train[:3])
print("\nX_test (scaled):")
print(X_test)
```

**Giải thích**:
- ⚠️ **QUAN TRỌNG**:
  - `fit_transform()` trên **train** (học mean, std)
  - `transform()` trên **test** (dùng mean, std từ train)
- Tránh **data leakage**!

### Bước 7: Verify kết quả

```python
# Summary
print("\n" + "="*50)
print("PREPROCESSING SUMMARY")
print("="*50)
print(f"Original dataset: {len(dataset)} samples, {dataset.shape[1]} features")
print(f"After dropping ID & Name: {X.shape[1]} features")
print(f"Missing values: Imputed with mean")
print(f"Categorical encoding: Gender → 0/1")
print(f"Train set: {X_train.shape[0]} samples")
print(f"Test set: {X_test.shape[0]} samples")
print(f"Feature scaling: StandardScaler applied")
print("="*50)
```

---

## ✅ Complete Solution Code

```python
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.impute import SimpleImputer

# 1. Create/Load data
data = {
    'StudentID': [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    'Name': ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack'],
    'Gender': ['Female', 'Male', 'Male', 'Female', 'Female', 'Male', 'Female', 'Male', 'Female', 'Male'],
    'Age': [20, 19, 21, 20, 22, 19, 21, 20, 19, 22],
    'StudyHours': [5.5, np.nan, 3.2, 6.8, 2.1, 4.5, 7.2, 3.8, 5.0, np.nan],
    'PreviousScore': [85, 75, np.nan, 90, 65, 80, 95, 70, 88, 78],
    'AttendanceRate': [90, 80, 70, 95, np.nan, 85, 92, 75, np.nan, 88],
    'Passed': [1, 1, 0, 1, 0, 1, 1, 0, 1, 1]
}
dataset = pd.DataFrame(data)

# 2. Drop unnecessary columns
dataset = dataset.drop(['StudentID', 'Name'], axis=1)
X = dataset.drop('Passed', axis=1).values
y = dataset['Passed'].values

# 3. Handle missing values
imputer = SimpleImputer(missing_values=np.nan, strategy='mean')
X[:, 1:5] = imputer.fit_transform(X[:, 1:5])

# 4. Encode categorical
le = LabelEncoder()
X[:, 0] = le.fit_transform(X[:, 0])

# 5. Split
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42
)

# 6. Scale
sc = StandardScaler()
X_train = sc.fit_transform(X_train)
X_test = sc.transform(X_test)

# 7. Results
print(f"X_train shape: {X_train.shape}")
print(f"X_test shape: {X_test.shape}")
print(f"y_train shape: {y_train.shape}")
print(f"y_test shape: {y_test.shape}")
```

---

## 🎯 Expected Output

```
X_train shape: (8, 5)
X_test shape: (2, 5)
y_train shape: (8,)
y_test shape: (2,)
```

---

## 🚀 Extension Challenges

1. **Thử strategies khác**:
   ```python
   imputer = SimpleImputer(strategy='median')  # Thay vì mean
   ```

2. **One-Hot Encoding** thay vì Label Encoding:
   ```python
   from sklearn.preprocessing import OneHotEncoder
   from sklearn.compose import ColumnTransformer

   ct = ColumnTransformer(
       transformers=[('encoder', OneHotEncoder(), [0])],
       remainder='passthrough'
   )
   X = ct.fit_transform(X)
   ```

3. **Add more features**: Tạo feature mới như `StudyEfficiency = PreviousScore / StudyHours`

4. **Visualize** missing values:
   ```python
   import matplotlib.pyplot as plt
   import seaborn as sns

   sns.heatmap(dataset.isnull(), cbar=False)
   plt.show()
   ```

5. **Save preprocessed data**:
   ```python
   import joblib

   joblib.dump(sc, 'scaler.pkl')
   joblib.dump(le, 'label_encoder.pkl')
   np.save('X_train.npy', X_train)
   np.save('y_train.npy', y_train)
   ```

---

## 📚 Key Takeaways

1. ✅ **Luôn drop** columns không phải features (ID, Name)
2. ✅ **Impute missing values** trước khi encode/scale
3. ✅ **Encode categorical** trước khi scale
4. ✅ **Split TRƯỚC khi scale** để tránh data leakage
5. ✅ **fit_transform() trên train**, **transform() trên test**
6. ✅ **Save transformers** (scaler, encoder) để dùng cho production

---

## 🔗 Next Steps

Sau khi preprocessing, data này sẵn sàng cho:
- **Classification**: Logistic Regression, SVM, Random Forest (dự đoán Passed)
- **Model Evaluation**: Confusion Matrix, Accuracy
- Xem bài tập tiếp theo: [1-regressions/EXERCISE.md](../1-regressions/EXERCISE.md)
