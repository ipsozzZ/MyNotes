from transformers import GPTJForSequenceClassification, Trainer, TrainingArguments
import torch
 
# 加载预训练模型和分词器
model = GPTJForSequenceClassification.from_pretrained("gpt-j-6b", num_labels=2)  # 根据任务调整 num_labels
tokenizer = AutoTokenizer.from_pretrained("gpt-j-6b")
 
# 数据预处理函数
def tokenize_function(examples):
    return tokenizer(examples['text'], padding='max_length', truncation=True)
 
# 应用预处理函数到数据集
tokenized_datasets = dataset.map(tokenize_function, batched=True)
 
# 设置训练参数
training_args = TrainingArguments(
    output_dir='./results',          # 输出目录
    num_train_epochs=3,              # 训练周期数
    per_device_train_batch_size=1,   # 每设备训练批次大小
    warmup_steps=500,                # 热身步骤数
    weight_decay=0.01,               # 权重衰减率
    logging_dir='./logs',            # 日志目录
    logging_steps=10,                # 日志间隔步数
)
 
# 初始化 Trainer 类并开始训练
trainer = Trainer(
    model=model,                         # 要训练的模型
    args=training_args,                 # 训练参数
    train_dataset=tokenized_datasets['train'],  # 训练数据集
    eval_dataset=tokenized_datasets['test'],    # 验证数据集（如果存在）
)
 
trainer.train()  # 开始训练过程